package main

// The Webserver code base is copied from the go gitlab library examples, will rewrite so skipped tests

import (
	"github.com/xanzy/go-gitlab"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testPayload(fixture int) *gitlab.MergeEvent {
	var fixturePath string
	var err error
	switch fixture {
	case 1:
		fixturePath = "./tests/fixtures/merge_request_events/success.json"
	case 2:
		fixturePath = "./tests/fixtures/merge_request_events/cannot-be-merged.json"
	case 3:
		fixturePath = "./tests/fixtures/merge_request_events/merge-action.json"
	case 4:
		fixturePath = "./tests/fixtures/merge_request_events/approved-action.json"
	}

	if err != nil {
		return nil
	}

	jsonFile, err := os.Open(fixturePath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var gmre *gitlab.MergeEvent
	err = json.Unmarshal(byteValue, &gmre)
	if err != nil {
		fmt.Println("failed to Unmarshal test data")
	}
	return gmre
}

func TestHandleMRRequest(t *testing.T) {
	type test struct {
		fixture         int
		GitlabBotUserID int
		result          string
		err             error
	}
	tests := []test{
		{1, 99999, "successfully added merge request to processing queue.", nil},
		{2, 99999, "ignoring as merge request cannot be merged.", nil},
		{3, 99999, "ignoring approved/merge action.", nil},
		{4, 99999, "ignoring approved/merge action.", nil},
		{1, 1, "ignoring request initiated through this service.", nil},
	}

	// Do not handle the channel element of this at the moment
	requests := make(chan MergeRequest, 1)

	for _, tc := range tests {
		event := testPayload(tc.fixture)

		wh := webhook{
			Secret:          "test",
			EventsToAccept:  []gitlab.EventType{gitlab.EventTypeMergeRequest},
			GitlabBotUserID: tc.GitlabBotUserID,
			Requests:        requests,
		}

		got, err := wh.handleMRRequest(event)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tc.result, got)
	}
}
