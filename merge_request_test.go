package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xanzy/go-gitlab"
)

// Setup

type mockGitlab struct {
	GitlabWrapper
	mock.Mock
}

func (o *mockGitlab) GetMergeRequest(pid interface{}, mergeRequest int, opt *gitlab.GetMergeRequestsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
	var fixturePath string
	var err error
	var http_response *http.Response
	switch mergeRequest {
	case 1:
		fixturePath = "./tests/fixtures/merge_requests/assigned-mr.json"
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	case 2:
		err = fmt.Errorf("failed to get mr: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", pid, mergeRequest)
		http_response = &http.Response{StatusCode: http.StatusNotFound}
	case 3:
		err = fmt.Errorf("failed to get mr: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", pid, mergeRequest)
		http_response = &http.Response{StatusCode: http.StatusNotFound}
	}

	r := &gitlab.Response{
		Response:     http_response,
		TotalItems:   0,
		TotalPages:   0,
		ItemsPerPage: 0,
		CurrentPage:  0,
		NextPage:     0,
		PreviousPage: 0,
	}

	if err != nil {
		return nil, r, err
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
		fmt.Println("error encountered unmarshalling test data!")
	}

	return gmr, r, nil
}

func (o *mockGitlab) CurrentUser(options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
	return nil, nil, nil
}

func (o *mockGitlab) GetConfiguration(pid interface{}, mr int, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequestApprovals, *gitlab.Response, error) {
	var fixturePath string
	var err error
	var http_response *http.Response
	switch mr {
	case 1:
		fixturePath = "./tests/fixtures/merge_request_approvals/0-suggested-1-approval.json"
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	case 2:
		fixturePath = "./tests/fixtures/merge_request_approvals/1-suggested-1-approval.json"
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	case 3:
		fixturePath = "./tests/fixtures/merge_request_approvals/3-suggested-1-approval.json"
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	case 4:
		fixturePath = "./tests/fixtures/merge_request_approvals/3-suggested-4-approval.json"
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	case 5:
		err = fmt.Errorf("failed to get approvers: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", pid, mr)
		http_response = &http.Response{StatusCode: http.StatusNotFound}
	case 6:
		err = fmt.Errorf("failed to get approvers: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", pid, mr)
		http_response = &http.Response{StatusCode: http.StatusNotFound}
	case 7:
		fixturePath = "./tests/fixtures/merge_request_approvals/1-suggested-0-approval.json"
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	}

	r := &gitlab.Response{
		Response:     http_response,
		TotalItems:   0,
		TotalPages:   0,
		ItemsPerPage: 0,
		CurrentPage:  0,
		NextPage:     0,
		PreviousPage: 0,
	}

	if err != nil {
		return nil, r, err
	}

	jsonFile, err := os.Open(fixturePath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var gmra *gitlab.MergeRequestApprovals

	err = json.Unmarshal(byteValue, &gmra)
	if err != nil {
		fmt.Println("error encountered unmarshalling test data!")
	}

	return gmra, r, nil
}

func (o *mockGitlab) UpdateMergeRequest(pid interface{}, mergeRequest int, opt *gitlab.UpdateMergeRequestOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
	var err error
	var http_response *http.Response
	switch mergeRequest {
	// Successfully remove an existing or unexisting Reviewer on a MR
	case 1:
		err = nil
		http_response = &http.Response{StatusCode: http.StatusAccepted}
	// Project does not exist
	case 2:
		err = fmt.Errorf("failed to assign reviewer on project: PUT https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", pid, mergeRequest)
		http_response = &http.Response{StatusCode: http.StatusNotFound}
	// MergeRequest does not exist
	case 3:
		err = fmt.Errorf("failed to assign reviewer on project: PUT https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", pid, mergeRequest)
		http_response = &http.Response{StatusCode: http.StatusNotFound}
	}

	r := &gitlab.Response{
		Response:     http_response,
		TotalItems:   0,
		TotalPages:   0,
		ItemsPerPage: 0,
		CurrentPage:  0,
		NextPage:     0,
		PreviousPage: 0,
	}

	return nil, r, err
}

// Setup

func TestMain(m *testing.M) {
	// Do not output log entries when running tests
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

// Tests

func TestUnsetMRReviewer(t *testing.T) {
	type test struct {
		mergeReqID int
		err        error
	}

	tests := []test{
		// successful unset of reviewer (if one assigned or not)
		{mergeReqID: 1, err: nil},
		// unsuccessful due to project not existing
		{mergeReqID: 2, err: fmt.Errorf("failed to assign reviewer on project: PUT https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", 1, 2)},
		// unsuccessful due to merge request not existing
		{mergeReqID: 3, err: fmt.Errorf("failed to assign reviewer on project: PUT https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", 1, 3)},
	}

	m := &mockGitlab{}

	for _, tc := range tests {
		mr := MergeRequest{
			pathWithNamespace: "test/test",
			group:             "test",
			projectID:         1,
			mergeReqID:        tc.mergeReqID,
		}

		err := mr.unsetMRReviwer(m)
		if err != nil {
			assert.Contains(t, err.Error(), tc.err.Error())
			assert.Contains(t, err.Error(), "404 {message: 404")
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestSetMRReviewer(t *testing.T) {
	type test struct {
		mergeReqID int
		err        error
	}

	tests := []test{
		// successful set reviewer (if one assigned (will overwrite) or not)
		{mergeReqID: 1, err: nil},
		// unsuccessful due to project not existing
		{mergeReqID: 2, err: fmt.Errorf("failed to assign reviewer on project: PUT https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", 1, 2)},
		// unsuccessful due to merge request not existing
		{mergeReqID: 3, err: fmt.Errorf("failed to assign reviewer on project: PUT https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", 1, 3)},
	}

	m := &mockGitlab{}
	reviewer1 := &gitlab.BasicUser{
		ID:       1,
		Name:     "Test User 1",
		Username: "test.user1",
	}
	reviewer2 := &gitlab.BasicUser{
		ID:       2,
		Name:     "Test User 2",
		Username: "test.user2",
	}
	reviewerIDs := []*gitlab.BasicUser{reviewer1, reviewer2}

	for _, tc := range tests {
		mr := MergeRequest{
			pathWithNamespace: "test/test",
			group:             "test",
			projectID:         1,
			mergeReqID:        tc.mergeReqID,
		}

		err := mr.setMRReviwer(m, reviewerIDs)
		if err != nil {
			assert.Contains(t, err.Error(), tc.err.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestGetMRApprovers(t *testing.T) {
	type test struct {
		mergeReqID        int
		approversReturned int
		approvalsRequired int
		validIDs          []int
		err               error
	}

	tests := []test{
		// success: no suggested approvers
		{mergeReqID: 1, approversReturned: 0, approvalsRequired: 1, validIDs: []int{}, err: errors.New("no suggested approvers.")},
		// success: 1 suggested approver, 1 approval required
		{mergeReqID: 2, approversReturned: 1, approvalsRequired: 1, validIDs: []int{1}, err: nil},
		// success: multiple suggested approvers, 1 approval required
		{mergeReqID: 3, approversReturned: 3, approvalsRequired: 1, validIDs: []int{1, 2, 3}, err: nil},
		// success: multiple suggested approvers, more approvals required than available
		{mergeReqID: 4, approversReturned: 3, approvalsRequired: 4, validIDs: []int{1, 2, 3}, err: nil},
		// unsuccessful due to project not existing
		{mergeReqID: 5, err: fmt.Errorf("failed to get approvers: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", 1, 5)},
		// unsuccessful due to merge request not existing
		{mergeReqID: 6, err: fmt.Errorf("failed to get approvers: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", 1, 6)},
		// unsuccessful: no approvals required
		{mergeReqID: 7, approversReturned: 3, approvalsRequired: 0, validIDs: []int{}, err: errors.New("approvals required is zero, will not assign reviewer.")},
	}

	m := &mockGitlab{}

	for _, tc := range tests {
		mr := MergeRequest{
			pathWithNamespace: "test/test",
			group:             "test",
			projectID:         1,
			mergeReqID:        tc.mergeReqID,
		}

		// Calls mock function (above): GetConfiguration which returns fixture merge requests
		approvers, approvalsRequired, err := mr.getMRApprovers(m)
		if err != nil {
			assert.Contains(t, err.Error(), tc.err.Error())
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, tc.approvalsRequired, approvalsRequired)
		assert.Len(t, approvers, tc.approversReturned)

		for _, approver := range approvers {
			assert.Contains(t, tc.validIDs, approver.ID)
		}
	}
}

func TestGetMR(t *testing.T) {
	type test struct {
		mergeReqID int
		err        error
	}

	tests := []test{
		// successful Get MR
		{mergeReqID: 1, err: nil},
		// unsuccessful due to project not existing
		{mergeReqID: 2, err: fmt.Errorf("failed to get mr: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Project Not Found}", 1, 2)},
		// unsuccessful due to merge request not existing
		{mergeReqID: 3, err: fmt.Errorf("failed to get mr: GET https://gitlab.local/api/v4/projects/%d/merge_requests/%d: 404 {message: 404 Not Found}", 1, 3)},
	}

	m := &mockGitlab{}

	for _, tc := range tests {
		mr := MergeRequest{
			pathWithNamespace: "test/test",
			group:             "test",
			projectID:         1,
			mergeReqID:        tc.mergeReqID,
		}

		err, res := mr.getMR(m)
		if err != nil {
			assert.Contains(t, err.Error(), tc.err.Error())
			assert.Contains(t, err.Error(), "404 {message: 404")
			continue
		}
		assert.NoError(t, err)
		assert.Len(t, res.Reviewers, 1)
	}
}

func TestGroupPath(t *testing.T) {
	type test struct {
		namespaceWithPath string
		want              string
		err               string
	}

	tests := []test{
		{namespaceWithPath: "test/test", want: "test", err: ""},
		{namespaceWithPath: "test/test/test", want: "test/test", err: ""},
		{namespaceWithPath: "test", want: "test", err: "path does not contain /."},
	}

	for _, tc := range tests {
		compare, err := groupPath(tc.namespaceWithPath)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err)
		}

		assert.Equal(t, compare, tc.want)
	}
}
