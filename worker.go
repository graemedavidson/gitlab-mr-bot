package main

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type WorkerStatus uint8

const (
	WorkerWorking WorkerStatus = iota
	WorkerWaiting
)

type Worker struct{}

func NewWorker() *Worker {
	return &Worker{}
}

// Working routing to handle assigning Reviewers to MergeRequests asynchronously
func (w *Worker) Run(requests chan MergeRequest, responses chan MRResponse, status chan WorkerStatus, gitClient GitlabWrapper, slack SlackWrapper, config Config, cache *localCache) {
	for mergeRequestJob := range requests {
		logger := log.WithFields(log.Fields{"group": mergeRequestJob.Group(), "project_id": mergeRequestJob.ProjectID(), "merge_request_id": mergeRequestJob.MergeReqID()})

		status <- WorkerWorking

		logger.Debug("processing mr to assign reviewer.")
		resultMessage, err := w.ProcessMR(gitClient, mergeRequestJob, slack, config, responses, cache)
		if err != nil {
			logger.Error(err.Error())
		} else {
			logger.Info(resultMessage)
		}
		responses <- MRResponse{status: resultMessage, err: err}

		status <- WorkerWaiting
	}
}

// Checks for current reviews and if none, assigns randomly from suggested approvers
//gocyclo:ignore
func (w *Worker) ProcessMR(gitClient GitlabWrapper, mr MergeRequests, slack SlackWrapper, config Config, responses chan MRResponse, cache *localCache) (string, error) {
	logger := log.WithFields(log.Fields{"group": mr.Group(), "project_id": mr.ProjectID(), "merge_request_id": mr.MergeReqID()})

	promProcessedMRs.WithLabelValues(mr.Group()).Inc()

	err, mrResult := mr.getMR(gitClient)
	if err != nil {
		return "", err
	}

	if mr.WorkInProgress() {
		if len(mrResult.Reviewers) > 0 {
			err = mr.unsetMRReviwer(gitClient)
			if err != nil {
				return "", err
			}
			promRemoveReviewer.WithLabelValues(mr.Group()).Inc()
			return "mr set to wip, un-assigned reviewer.", nil
		}

		return "mr set to wip, no action required.", nil
	}

	if len(mrResult.Reviewers) > 0 {
		promIgnoreActions.WithLabelValues("reviewer_already_assigned", mr.Group()).Inc()
		return "reviewer already assigned.", nil
	}

	approvers, approvalsRequired, err := mr.getMRApprovers(gitClient)
	if err != nil {
		return "", err
	}

	slackChannel, slackChannelID, err := getSlackChannel(mr.PathWithNamespace(), config.GroupChannels)
	if err != nil {
		return "", err
	}

	// Check for any missing usernames from the cache in comparison to the codeowners
	// if missing the system will not know the slack user ID which is required for requesting
	// the slack user status, therefore we must grab it via the allocated channel for sending slack
	// messages
	var usernames []string
	for _, user := range approvers {
		usernames = append(usernames, user.Username)
	}
	missingUsernames := cache.getMissingIDs(usernames...)
	if len(missingUsernames) > 0 {
		slackUsernames, err := getSlackUserIDs(slack, cache, slackChannelID, mr)
		if err != nil {
			return "", err
		}
		logger.WithFields(log.Fields{"missing_ids": slackUsernames}).Debug("missing cache entries, fetching usernames from channel")
		// We have pulled out all the userids but with no way of knowing what ID belongs to who, so call
		// a full update cache for all users in the channel. This is initially expensive but should mean that
		// we gain majority coverage of all git username to slack username and slack id quickly and the IDS
		// will remain mapped in memory in the cache.

		// ToDo: Handle a large number of user ids pulled from the slack channel
		err = updateCache(slack, cache, mr, slackUsernames, config)
		if err != nil {
			return "", err
		}
	}

	approvers = checkCache(slack, cache, approvers, mr, config)

	if len(approvers) == 0 {
		promIgnoreActions.WithLabelValues("no_available_approvers", mr.Group()).Inc()
		return "", errors.New("no approvers available after slack status checks.")
	}

	selectedApprovers := selectApprovers(approvers, approvalsRequired)
	logger.WithFields(log.Fields{"selected": selectedApprovers, "approvals_required": approvalsRequired, "num_approvers": len(approvers)}).Debug("selected to assign to mr.")

	err = mr.setMRReviwer(gitClient, selectedApprovers)
	if err != nil {
		return "", err
	}

	if slack.webhookURLSet() {
		logger.WithFields(log.Fields{"channel": slackChannel}).Debug("send slack message.")
		err = sendSlackMsg(slack, slackChannel, selectedApprovers, mr)
		if err != nil {
			return "", err
		}
	} else {
		logger.WithFields(log.Fields{"group": mr.Group()}).Warn("no slack channel configured for group.")
		promSlackMsgsErrors.WithLabelValues("no_slack_channel_configured", mr.Group(), "").Inc()
	}

	return "successfully processed merge request.", nil
}

// checkCache: check the cache for user status and update where required, then pass on a list of available approvers
func checkCache(slack SlackWrapper, cache *localCache, suggestedApprovers []*gitlab.BasicUser, mr MergeRequests, config Config) []*gitlab.BasicUser {
	logger := log.WithFields(log.Fields{"group": mr.Group(), "project_id": mr.ProjectID(), "merge_request_id": mr.MergeReqID()})

	// var unexpectedErrors []string
	var approvers []*gitlab.BasicUser

	for _, gitUser := range suggestedApprovers {
		cachedUser, err := cache.read(gitUser.Username)
		if err != nil {
			switch err.Error() {
			case "no_user_in_cache":
				// this should not happen as missing users are checked before this (above) and should be added before this check function.
				promErrors.WithLabelValues("user_not_found_in_cache").Inc()
				logger.WithFields(log.Fields{"error": err}).Error("user not found in cache.")
				continue
			case "user_data_expired":
				slackUsersData, _, err := getUsersInfo(slack, cachedUser.slackUserID)
				if err != nil {
					logger.WithFields(log.Fields{"error": err}).Error("failed to get slack user data.")
					// do not block on error which is hopefully temporary, continue to review other users
					continue
				}
				for _, s := range *slackUsersData {
					u := userMeta{username: cachedUser.username, slackUserID: cachedUser.slackUserID, status: strings.ToLower(s.Profile.StatusText)}
					t := time.Now()
					ttl := getStatusTTL(config.UserStatuses, s.Profile.StatusText)
					expire := t.Add(time.Hour * time.Duration(ttl))
					cache.update(u, expire.Unix())
				}
			default:
				// will NOT execute because of the line preceding the switch.
			}
		}
		if cachedUser.status == "out sick" || cachedUser.status == "vacationing" || cachedUser.status == "holiday" {
			promSlackStatusUnavailable.WithLabelValues(cachedUser.status, mr.Group()).Inc()
			logger.WithFields(log.Fields{"reason": cachedUser.status, "username": gitUser.Username}).Debug("user unavailable due to slack status.")
		} else {
			approvers = append(approvers, gitUser)
		}
	}
	return approvers
}

// updateCache: Update the user cache for all passed updates (list of usernames)
func updateCache(slack SlackWrapper, cache *localCache, mr MergeRequests, updates []string, config Config) error {
	logger := log.WithFields(log.Fields{"group": mr.Group(), "project_id": mr.ProjectID(), "merge_request_id": mr.MergeReqID()})

	logger.WithFields(log.Fields{"num_updates": len(updates)}).Debug("updating cache entries.")

	slackUsersData, missingIDs, err := getUsersInfo(slack, updates...)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("failed to get slack user data.")
		return err
	}

	// Gitlab Users that do not exist in Slack
	if len(missingIDs) > 0 {
		logger.WithFields(log.Fields{"num_missing": len(missingIDs)}).Debug("gitlab users in codeowners without matching user in slack channel.")
		promSlackUsersMissing.WithLabelValues(mr.Group()).Inc()
	}

	if len(*slackUsersData) > 0 {
		for _, s := range *slackUsersData {
			logger.WithFields(log.Fields{"username": len(s.Name)}).Debug("updating cache entry.")
			u := userMeta{username: s.Name, slackUserID: s.ID, status: strings.ToLower(s.Profile.StatusText)}
			t := time.Now()
			ttl := getStatusTTL(config.UserStatuses, s.Profile.StatusText)
			expire := t.Add(time.Hour * time.Duration(ttl))
			cache.update(u, expire.Unix())
		}
	}

	return nil
}

// getSlackUserIDs: get list of slack user ids from channel set in configuration and compare against ids found in the
// local cache returning a list of all missing ids.
func getSlackUserIDs(slack SlackWrapper, cache *localCache, slackChannelID string, mr MergeRequests) ([]string, error) {
	slackUsers, err := getUsersInConversation(slack, slackChannelID)
	if err != nil {
		return nil, err
	}
	return slackUsers, nil
}

// Select random reviewers upto approvals required
func selectApprovers(approvers []*gitlab.BasicUser, approvalsRequired int) []*gitlab.BasicUser {
	var selectedApprovers []*gitlab.BasicUser

	if approvalsRequired > len(approvers) {
		return approvers
	}

	rand.Seed(time.Now().UnixNano())
	// Random is by the index of the result.
	randomApprovalIndexes := rand.Perm(len(approvers))

	for i := 0; i < approvalsRequired; i++ {
		randID := randomApprovalIndexes[i]
		selectedApprovers = append(selectedApprovers, approvers[randID])
	}

	return selectedApprovers
}

// getSlackChannel: return slack channel and id to post to based on the group set in the MR payload
func getSlackChannel(pathWithNamespace string, groupChannels map[string]GroupChannel) (string, string, error) {
	compare := strings.ToLower(pathWithNamespace)
	var err error
	for {
		if channel, ok := groupChannels[compare]; ok {
			return channel.SlackChannel, channel.SlackChannelID, nil
		}

		compare, err = groupPath(compare)
		if err != nil {
			return "", "", errors.New("no slack channel configured.")
		}
	}
}

// getStatusTTL: return the ttl in hours assigned to a status
// TODO: make the statuses generic, instead of being allocated to slack. The statuses themselves have
//       nothing to do with slack and could come from any place.
func getStatusTTL(userStatuses map[string]int, status string) int {
	logger := log.WithFields(log.Fields{"status": status})
	for k, v := range userStatuses {
		if k == strings.ToLower(status) {
			return v
		}
	}
	logger.Debug("returned user status has no config entry, using default.")
	return userStatuses[""]
}
