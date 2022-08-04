package main

// ToDo: Productionise the WebServer Config which is currently taken from:
// https://github.com/xanzy/go-gitlab/blob/master/examples/webhook.go

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type webhook struct {
	Secret          string
	EventsToAccept  []gitlab.EventType
	GitlabBotUserID int
	Requests        chan MergeRequest
}

// Handle the different types of requests/gitlab events
func (hook webhook) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	event, err := hook.parse(request)
	if err != nil {
		// TODO: Add prom metrics
		log.WithFields(log.Fields{"error": err}).Error("could not parse the webhook event.")
		writer.WriteHeader(500)
		_, err := writer.Write([]byte(fmt.Sprintf("could not parse the webhook event: %v", err)))
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to write fail header to external connection.")
		}
		return
	}

	switch v := event.(type) {
	default:
		promErrors.WithLabelValues("unexpected_event_type").Inc()
		log.WithFields(log.Fields{"event_type": v}).Error("unexpected type found in payload.")
	case *gitlab.MergeEvent:

		// type assertion when passing to func
		res, err := hook.handleMRRequest(event.(*gitlab.MergeEvent))
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("handling MergeEvent request.")
			writer.WriteHeader(500)
			_, err := writer.Write([]byte(fmt.Sprintf("error handling the event: %v", err)))
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("failed to write fail header to external connection.")
			}
			return
		}

		writer.WriteHeader(202)
		_, err = writer.Write([]byte(fmt.Sprintf("%v", res)))
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to write fail header to external connection.")
		}
	}
}

// Handle a MergerRequest Gitlab Event
// Currently never returns an error.
func (hook webhook) handleMRRequest(event *gitlab.MergeEvent) (string, error) {
	// strip project name from path
	groupPath, _ := groupPath(event.Project.PathWithNamespace)
	mr := MergeRequest{
		pathWithNamespace: event.Project.PathWithNamespace,
		group:             groupPath,
		projectID:         event.Project.ID,
		projectName:       event.Project.Name,
		projectWebURL:     event.Project.WebURL,
		mergeReqID:        event.ObjectAttributes.IID,
		mergeReqURL:       event.ObjectAttributes.URL,
		mergeReqTitle:     event.ObjectAttributes.Title,
		workInProgress:    event.ObjectAttributes.WorkInProgress,
	}

	logger := log.WithFields(log.Fields{"group": mr.group, "project_id": mr.projectID, "merge_request_id": mr.mergeReqID})
	logger.Debug("handling mr event.")
	promEvents.WithLabelValues("merge_event", mr.group).Inc()

	// Ignore events which originate from this service as they are calls made when the service
	// updates the MergeRequest which then calls the service again.
	if event.User.ID == hook.GitlabBotUserID {
		promRecursiveCalls.WithLabelValues(mr.group).Inc()
		logger.WithFields(log.Fields{"bot_username": event.User.Username, "bot_id": event.User.ID}).Debug("ignoring request initiated through this service.")
		// not an error
		return "ignoring request initiated through this service.", nil
	}

	// Ignore actions which should mean a reviewer should not be set
	result, _ := regexp.MatchString("(approved|merge)", strings.ToLower(event.ObjectAttributes.Action))
	if result {
		promIgnoreActions.WithLabelValues(event.ObjectAttributes.Action, mr.group).Inc()
		logger.Debug("ignoring approved/merge action.")
		return "ignoring approved/merge action.", nil
	}

	if event.ObjectAttributes.MergeStatus == "cannot_be_merged" {
		promIgnoreActions.WithLabelValues("mr_cannot_merge", mr.group).Inc()
		logger.Debug("ignoring as merge request cannot be merged.")
		return "ignoring as merge request cannot be merged.", nil
	}

	if event.ObjectAttributes.WorkInProgress {
		promIgnoreActions.WithLabelValues("mr_is_wip", mr.group).Inc()
		logger.Debug("merge request is a wip. continue to remove any assigned reviewers.")
		// A merge request event does not include the list of reviewers (if any) assigned. Therefore it must be passed on to the
		// queue and processed by getting the MR to determine if a reviewer is already set and therefore should be removed.
		// Hopefully future updates will include this in the event and it can be handled here.
		// return "Ignoring as merge request is work in progress", nil
	}

	// MR has passed checks; assign reviewer asynchronously
	hook.Requests <- mr

	return "successfully added merge request to processing queue.", nil
}

// parse verifies and parses the events specified in the request and returns the parsed event or an error.
func (hook webhook) parse(r *http.Request) (interface{}, error) {
	defer func() {
		if _, err := io.Copy(ioutil.Discard, r.Body); err != nil {
			promErrors.WithLabelValues("discard_event_body").Inc()
			log.WithFields(log.Fields{"error": err}).Error("could not discard request body.")
		}
		if err := r.Body.Close(); err != nil {
			promErrors.WithLabelValues("close_request_body").Inc()
			log.WithFields(log.Fields{"error": err}).Error("could not close request body.")
		}
	}()

	if r.Method != http.MethodPost {
		promErrors.WithLabelValues("invalid_http_method").Inc()
		return nil, errors.New("invalid http method")
	}

	if len(hook.Secret) > 0 {
		signature := r.Header.Get("X-Gitlab-Token")
		if signature != hook.Secret {
			promErrors.WithLabelValues("token_validation").Inc()
			return nil, errors.New("token validation failed")
		}
	}

	event := r.Header.Get("X-Gitlab-Event")
	if strings.TrimSpace(event) == "" {
		promErrors.WithLabelValues("x_gitlab_event_header").Inc()
		return nil, errors.New("missing X-Gitlab-Event Header")
	}

	eventType := gitlab.EventType(event)
	if !isEventSubscribed(eventType, hook.EventsToAccept) {
		promErrors.WithLabelValues("unsubscribed_event").Inc()
		return nil, errors.New("event not defined to be parsed")
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		promErrors.WithLabelValues("read_request_body").Inc()
		return nil, errors.New("error reading request body")
	}

	return gitlab.ParseWebhook(eventType, payload)
}

// Check an event type is expected
func isEventSubscribed(event gitlab.EventType, events []gitlab.EventType) bool {
	for _, e := range events {
		if event == e {
			return true
		}
	}
	return false
}
