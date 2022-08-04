package main

import (
	"errors"

	"github.com/slack-go/slack"
	"github.com/xanzy/go-gitlab"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Setup

type mockSlack struct {
	SlackWrapper
	mock.Mock
	wh_url string
}

func (s *mockSlack) PostWebhook(msg *slack.WebhookMessage) error {
	if s.wh_url == "fail" {
		return errors.New("failed to send slack message!")
	}
	return nil
}

// Tests

func TestSendSlackMsg(t *testing.T) {

	type test struct {
		url string
	}

	tests := []test{
		{url: "pass"},
		{url: "fail"},
	}

	reviewer1 := &gitlab.BasicUser{
		ID:       1,
		Name:     "Test User 1",
		Username: "test.user1",
	}
	reviewers := []*gitlab.BasicUser{
		reviewer1,
	}
	mr := MergeRequest{
		pathWithNamespace: "test/test",
		group:             "test",
		projectID:         1,
		mergeReqID:        1,
		mergeReqURL:       "https://gitlab.local/merge_requests/1",
		mergeReqTitle:     "test",
		workInProgress:    false,
	}

	for _, tc := range tests {
		ms := &mockSlack{wh_url: tc.url}

		err := sendSlackMsg(ms, "test", reviewers, mr)

		if err != nil {
			assert.Equal(t, err.Error(), "failed to send slack message!")
		} else {
			assert.NoError(t, err)
		}
	}
}
