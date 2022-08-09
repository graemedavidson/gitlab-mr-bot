package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/slack-go/slack"
	"github.com/xanzy/go-gitlab"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMergeRequest struct {
	pathWithNamespace string
	group             string
	projectID         int
	projectName       string
	projectWebURL     string
	mergeReqID        int
	mergeReqURL       string
	mergeReqTitle     string
	workInProgress    bool
}

func (mr MockMergeRequest) getMR(gc GitlabWrapper) (error, *gitlab.MergeRequest) {
	var fixturePath string
	var err error
	switch mr.mergeReqID {
	case 1:
		fixturePath = "./tests/fixtures/merge_requests/assigned-mr.json"
	case 2:
		fixturePath = "./tests/fixtures/merge_requests/unassigned-mr.json"
	case 3:
		fixturePath = "./tests/fixtures/merge_requests/assigned-wip.json"
	case 4:
		fixturePath = "./tests/fixtures/merge_requests/unassigned-wip.json"
	case 5:
		// Are these errors really worth it?
		err = fmt.Errorf("failed to get mr: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", mr.projectID, mr.mergeReqID)
	case 6:
		// Are these errors really worth it?
		err = fmt.Errorf("failed to get mr: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", mr.projectID, mr.mergeReqID)
	}

	if err != nil {
		return err, nil
	}

	jsonFile, err := os.Open(fixturePath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var gmr *gitlab.MergeRequest
	err = json.Unmarshal(byteValue, &gmr)
	if err != nil {
		fmt.Println("failed to unmarshal test data")
	}
	return nil, gmr
}

func (mr MockMergeRequest) getMRApprovers(gc GitlabWrapper) ([]*gitlab.BasicUser, int, error) {
	a1 := &gitlab.BasicUser{ID: 1, Name: "Test 1", Username: "test1"}
	a2 := &gitlab.BasicUser{ID: 2, Name: "Test 2", Username: "test2"}
	a3 := &gitlab.BasicUser{ID: 3, Name: "Test 3", Username: "test3"}
	approvers := []*gitlab.BasicUser{a1, a2, a3}
	return approvers, 1, nil
}

func (mr MockMergeRequest) setMRReviwer(gc GitlabWrapper, reviewers []*gitlab.BasicUser) error {
	return nil
}

func (mr MockMergeRequest) unsetMRReviwer(gc GitlabWrapper) error {
	return nil
}

func (mr MockMergeRequest) PathWithNamespace() string {
	return mr.pathWithNamespace
}

func (mr MockMergeRequest) Group() string {
	return mr.group
}

func (mr MockMergeRequest) ProjectID() int {
	return mr.projectID
}

func (mr MockMergeRequest) ProjectName() string {
	return mr.projectName
}

func (mr MockMergeRequest) ProjectWebURL() string {
	return mr.projectWebURL
}

func (mr MockMergeRequest) MergeReqID() int {
	return mr.mergeReqID
}

func (mr MockMergeRequest) MergeReqURL() string {
	return mr.mergeReqURL
}

func (mr MockMergeRequest) MergeReqTitle() string {
	return mr.mergeReqTitle
}

func (mr MockMergeRequest) WorkInProgress() bool {
	return mr.workInProgress
}

type MockSlack struct {
	mock.Mock
	wh_url string
}

func (s *MockSlack) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	return "", "", nil
}

func (s *MockSlack) GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error) {
	var users []string
	switch params.ChannelID {
	case "A":
		users = []string{"1", "2"}
	case "B":
		users = []string{"1", "2", "3"}
	case "C":
		users = []string{"1", "2"}
	case "D":
		users = []string{}
	case "E":
		return nil, "", errors.New("channel_not_found")
	}
	return users, "", nil
}

func (s *MockSlack) GetUsersInfo(users ...string) (*[]slack.User, error) {
	userInfo := []slack.User{}
	u1 := slack.User{
		ID:   "1",
		Name: "test1",
		Profile: slack.UserProfile{
			StatusText: "",
		},
	}
	u2 := slack.User{
		ID:   "2",
		Name: "test2",
		Profile: slack.UserProfile{
			StatusText: "",
		},
	}
	u3 := slack.User{
		ID:   "1",
		Name: "test1",
		Profile: slack.UserProfile{
			StatusText: "Out Sick",
		},
	}
	switch s.wh_url {
	case "A":
		userInfo = append(userInfo, u1)
	case "B":
		userInfo = append(userInfo, u3)
	case "C":
		userInfo = append(userInfo, u1, u2)
	case "E":
		return nil, errors.New("user_not_found")
	case "F":
		userInfo = append(userInfo, u1)
	case "G":
		userInfo = append(userInfo, u1)
	case "H":
		userInfo = append(userInfo, u1, u2)
	}
	return &userInfo, nil
}

// Tests:

func TestProcessMR(t *testing.T) {
	mockGitClient := &mockGitlab{}
	mockConfig := Config{
		GroupChannels: map[string]GroupChannel{
			"repo":           {SlackChannel: "channel", SlackChannelID: "AAAAA"},
			"sub-group/repo": {SlackChannel: "channel", SlackChannelID: "AAAAA"},
			"sub_group/repo": {SlackChannel: "channel", SlackChannelID: "AAAAA"},
			"test/test":      {SlackChannel: "channel", SlackChannelID: "AAAAA"},
		},
	}

	worker := NewWorker()
	mockResponses := make(chan MRResponse, 10000)

	cache := newLocalCache()
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	cache.update(cache1, timeExpire.Unix())
	cache.update(cache2, timeExpire.Unix())

	type test struct {
		result string
		err    error
		MRID   int
		WIP    bool
		wh_url string
		path   string
	}

	tests := []test{
		{"reviewer already assigned.", nil, 1, false, "", "test/test"},
		{"successfully processed merge request.", nil, 2, false, "", "test/test"},
		{"mr set to wip, un-assigned reviewer.", nil, 3, true, "", "test/test"},
		// unassigned wip is correct state and ignored - might want to add a different result message for this.
		{"mr set to wip, no action required.", nil, 4, true, "", "test/test"},
		{"", fmt.Errorf("no slack channel configured."), 2, false, "test", "fail/fail"},
	}

	for _, tc := range tests {
		mockSlack := MockSlack{wh_url: tc.wh_url}
		mockMR := MockMergeRequest{
			pathWithNamespace: tc.path,
			group:             "test",
			projectID:         1,
			mergeReqID:        tc.MRID,
			workInProgress:    tc.WIP,
		}

		got, err := worker.ProcessMR(mockGitClient, mockMR, &mockSlack, mockConfig, mockResponses, cache)

		if err != nil {
			assert.Equal(t, err, tc.err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, tc.result, got)
	}
}

func TestSelectApprovers(t *testing.T) {
	type test struct {
		approvers         []*gitlab.BasicUser
		approvalsRequired int
		validIDs          []int
	}

	a1 := &gitlab.BasicUser{
		ID:       1,
		Name:     "Test User 1",
		Username: "test.user1",
	}
	a2 := &gitlab.BasicUser{
		ID:       2,
		Name:     "Test User 2",
		Username: "test.user2",
	}
	a3 := &gitlab.BasicUser{
		ID:       3,
		Name:     "Test User 3",
		Username: "test.user3",
	}
	approvers := []*gitlab.BasicUser{a1, a2, a3}

	tests := []test{
		// should return empty list
		{approvers: approvers, approvalsRequired: 0, validIDs: []int{}},
		{approvers: approvers, approvalsRequired: 1, validIDs: []int{1, 2, 3}},
		{approvers: approvers, approvalsRequired: 2, validIDs: []int{1, 2, 3}},
		// should return 3 approvers
		{approvers: approvers, approvalsRequired: 4, validIDs: []int{1, 2, 3}},
	}

	for _, tc := range tests {
		selectedApprovers := selectApprovers(tc.approvers, tc.approvalsRequired)
		if len(tc.approvers) > tc.approvalsRequired {
			assert.Equal(t, tc.approvalsRequired, len(selectedApprovers))
		}
		for _, sa := range selectedApprovers {
			assert.Contains(t, tc.validIDs, sa.ID)
		}
	}
}

func TestGetSlackChannel(t *testing.T) {
	type test struct {
		search          string
		want_channel    string
		want_channel_id string
		err             string
	}

	mockChannelConfig := Config{
		GroupChannels: map[string]GroupChannel{
			"repo":           {SlackChannel: "channel", SlackChannelID: "AAAAA"},
			"sub-group/repo": {SlackChannel: "channel", SlackChannelID: "AAAAA"},
			"sub_group/repo": {SlackChannel: "channel", SlackChannelID: "AAAAA"},
		},
	}

	tests := []test{
		{search: "repo", want_channel: "channel", want_channel_id: "AAAAA", err: ""},
		{search: "sub-group/repo", want_channel: "channel", want_channel_id: "AAAAA", err: ""},
		{search: "sub_group/repo", want_channel: "channel", want_channel_id: "AAAAA", err: ""},
		{search: "SUB_GROUP/REPO", want_channel: "channel", want_channel_id: "AAAAA", err: ""},
		{search: "not-exist", want_channel: "", want_channel_id: "", err: "no slack channel configured."},
	}

	for _, tc := range tests {
		mr := MergeRequest{pathWithNamespace: tc.search}
		GroupChannels := mockChannelConfig.GroupChannels

		got_channel, got_channel_id, err := getSlackChannel(mr.pathWithNamespace, GroupChannels)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err)
		} else {
			assert.Equal(t, got_channel, tc.want_channel)
			assert.Equal(t, got_channel_id, tc.want_channel_id)
		}
	}
}

func TestCheckCache(t *testing.T) {
	mockConfig := Config{
		UserStatuses: map[string]int{
			"":           1,
			"out sick":   8,
			"vactioning": 8,
			"holiday":    8,
		},
	}

	mockMR := MockMergeRequest{
		pathWithNamespace: "test/test",
		group:             "test",
		projectID:         1,
		mergeReqID:        1,
		workInProgress:    false,
	}

	cache := newLocalCache()

	type test struct {
		seed               string
		suggestedApprovers []*gitlab.BasicUser
		approvers          []*gitlab.BasicUser
	}

	// Gitlab Users
	reviewer1 := &gitlab.BasicUser{
		ID:       1,
		Username: "test1",
	}
	reviewer2 := &gitlab.BasicUser{
		ID:       2,
		Username: "test2",
	}

	// Cache Users
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}

	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	timeExpired := timeNow.AddDate(-1, 0, 0)
	tests := []test{
		// Empty Cache: This should not happen for this function as in normal operation this is addressed, suggetes rewrite
		{
			"A",
			[]*gitlab.BasicUser{reviewer1},
			nil,
		},
		// User in cache with unexpired value, should be 1 to 1.
		{
			"F",
			[]*gitlab.BasicUser{reviewer1},
			[]*gitlab.BasicUser{reviewer1},
		},
		// User in cache with expired value, should be updated then 1 to 1.
		{
			"G",
			[]*gitlab.BasicUser{reviewer1},
			[]*gitlab.BasicUser{reviewer1},
		},
		// update multiple users in cache.
		{
			"H",
			[]*gitlab.BasicUser{reviewer1, reviewer2},
			[]*gitlab.BasicUser{reviewer1, reviewer2},
		},
	}

	for _, tc := range tests {
		// Use seed as method of passing a test param to the mock function
		mockSlack := MockSlack{wh_url: tc.seed}
		switch tc.seed {
		case "A":
			// Start with empty cache
		case "F":
			cache.update(cache1, timeExpire.Unix())
		case "G":
			cache.update(cache1, timeExpired.Unix())
		case "H":
			cache.update(cache1, timeExpired.Unix())
			cache.update(cache2, timeExpired.Unix())
		}

		got := checkCache(&mockSlack, cache, tc.suggestedApprovers, mockMR, mockConfig)

		assert.Equal(t, tc.approvers, got)
	}
}

func TestUpdateCache(t *testing.T) {
	// updte existing cache entry timestamp
	mockConfig := Config{
		UserStatuses: map[string]int{
			"":           1,
			"out sick":   8,
			"vactioning": 8,
			"holiday":    8,
		},
	}

	mockMR := MockMergeRequest{
		pathWithNamespace: "test/test",
		group:             "test",
		projectID:         1,
		mergeReqID:        1,
		workInProgress:    false,
	}

	cache := newLocalCache()

	type test struct {
		slackChanID string
		updates     []string
		got         map[string]cachedUser
		err         error
	}

	timeNow := time.Now()
	timeExpire1 := timeNow.Add(time.Hour * 1)
	timeExpire8 := timeNow.Add(time.Hour * 8)
	tests := []test{
		// Empty Cache
		{
			"A",
			[]string{"test1"},
			map[string]cachedUser{"test1": {user: userMeta{username: "test1", slackUserID: "1"}, expireTimestamp: timeExpire1.Unix()}},
			nil,
		},
		// Update Cache
		{
			"B",
			[]string{"test1"},
			map[string]cachedUser{"test1": {user: userMeta{username: "test1", slackUserID: "1", status: "out sick"}, expireTimestamp: timeExpire8.Unix()}},
			nil,
		},
		// Update Cache with new user
		{
			"C",
			[]string{"test1"},
			map[string]cachedUser{
				"test1": {user: userMeta{username: "test1", slackUserID: "1"}, expireTimestamp: timeExpire1.Unix()},
				"test2": {user: userMeta{username: "test2", slackUserID: "2"}, expireTimestamp: timeExpire1.Unix()},
			},
			nil,
		},
		// Error
		{
			"E",
			[]string{"test1"},
			nil,
			errors.New("slack: failed to get user details: user_not_found\n"),
		},
	}

	for _, tc := range tests {
		// Way of passing a test param to the mock function
		mockSlack := MockSlack{wh_url: tc.slackChanID}
		switch tc.slackChanID {
		case "A":
			// Start with empty cache
		case "B":
			u := userMeta{username: "test1", slackUserID: "1"}
			cache.update(u, timeExpire1.Unix())
		case "C":
			u := userMeta{username: "test1", slackUserID: "1"}
			cache.update(u, timeExpire1.Unix())
		}

		err := updateCache(&mockSlack, cache, mockMR, tc.updates, mockConfig)

		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
			continue
		}
		assert.NoError(t, err)

		for _, v := range tc.updates {
			assert.Equal(t, tc.got[v].user, cache.users[v].user)
			assert.WithinDuration(t, time.Unix(tc.got[v].expireTimestamp, 0), time.Unix(cache.users[v].expireTimestamp, 0), 0)
		}
	}

}

func TestGetSlackUserIDs(t *testing.T) {
	mockMR := MockMergeRequest{
		pathWithNamespace: "test/test",
		group:             "test",
		projectID:         1,
		mergeReqID:        1,
		workInProgress:    false,
	}

	mockSlack := MockSlack{}
	cache := newLocalCache()

	type test struct {
		slackChanID string
		got         []string
		err         error
	}

	tests := []test{
		{"A", []string{"1", "2"}, nil},
		// No users in slack channel?
		{"D", []string{}, nil},
		// channel does not exist
		{"E", nil, errors.New("slack: failed to get users in channel: channel_not_found\n")},
	}

	for _, tc := range tests {
		got, err := getSlackUserIDs(&mockSlack, cache, tc.slackChanID, mockMR)

		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, tc.got, got)
	}
}

func TestGetStatusTTL(t *testing.T) {
	mockConfig := Config{
		UserStatuses: map[string]int{
			"":           1,
			"out sick":   8,
			"vactioning": 8,
			"holiday":    8,
		},
	}

	type test struct {
		status   string
		expected int
	}

	tests := []test{
		{"", 1},
		{"out sick", 8},
		{"Out Sick", 8},
		{"OUT SICK", 8},
		{"vactioning", 8},
		{"holiday", 8},
		{"invalid", 1},
	}

	for _, tc := range tests {

		got := getStatusTTL(mockConfig.UserStatuses, tc.status)
		assert.Equal(t, tc.expected, got)
	}
}
