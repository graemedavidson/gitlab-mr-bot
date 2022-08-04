package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/xanzy/go-gitlab"
)

func getUsersInConversation(sw SlackWrapper, channelID string) ([]string, error) {
	promSlackAPIReqs.WithLabelValues("get_users_in_coversation").Inc()
	options := slack.GetUsersInConversationParameters{ChannelID: channelID}
	users, _, err := sw.GetUsersInConversation(&options)
	if err != nil {
		promSlackAPIErrs.WithLabelValues("get_users_in_coversation", err.Error()).Inc()
		return nil, fmt.Errorf("slack: failed to get users in channel: %s\n", err)
	}
	return users, nil
}

func getUsersInfo(sw SlackWrapper, users ...string) (*[]slack.User, []string, error) {
	promSlackAPIReqs.WithLabelValues("get_users_info").Inc()
	missingIDs := []string{}
	userInfo, err := sw.GetUsersInfo(users...)
	if err != nil {
		promSlackAPIErrs.WithLabelValues("get_users_info", err.Error()).Inc()
		return nil, nil, fmt.Errorf("slack: failed to get user details: %s\n", err)
	}

	// Returned user info is less than requested, determine missing.
	if len(users) != len(*userInfo) {
		for _, u := range users {
			for _, uinfo := range *userInfo {
				if u == uinfo.ID {
					continue
				}
				missingIDs = append(missingIDs, u)
			}
		}
	}

	return userInfo, missingIDs, nil
}

// Post message to slack channel via chat api
func sendSlackMsg(sw SlackWrapper, channel string, reviewers []*gitlab.BasicUser, mr MergeRequests) error {
	promSlackMsgs.WithLabelValues(mr.Group(), channel).Inc()

	var usernames []string
	for _, reviewer := range reviewers {
		usernames = append(usernames, reviewer.Username)
	}
	fmtUsernames := fmt.Sprintf("<@%s>", strings.Join(usernames, ">, <@"))

	attachment := slack.Attachment{
		Color:  "#1f81d1",
		Text:   fmt.Sprintf("%s you have been selected to review <%s|%s> in <%s|%s>", fmtUsernames, mr.MergeReqURL(), mr.MergeReqTitle(), mr.ProjectWebURL(), mr.ProjectName()),
		Footer: "Selections based on CODEOWNERS file",
	}

	_, _, err := sw.PostMessage(
		channel,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		promSlackMsgsErrors.WithLabelValues("msg_failed", mr.Group(), channel).Inc()
		return errors.New("failed to send slack message!")
	}
	return nil
}
