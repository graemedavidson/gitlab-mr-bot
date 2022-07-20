package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type MergeRequests interface {
	getMR(gc GitlabWrapper) (error, *gitlab.MergeRequest)
	getMRApprovers(gc GitlabWrapper) ([]*gitlab.BasicUser, int, error)
	setMRReviwer(gc GitlabWrapper, reviewers []*gitlab.BasicUser) error
	unsetMRReviwer(gc GitlabWrapper) error
	PathWithNamespace() string
	Group() string
	ProjectID() int
	ProjectName() string
	ProjectWebURL() string
	MergeReqID() int
	MergeReqURL() string
	MergeReqTitle() string
	WorkInProgress() bool
}

type MergeRequest struct {
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

// Accessors

func (mr MergeRequest) PathWithNamespace() string {
	return mr.pathWithNamespace
}

func (mr MergeRequest) Group() string {
	return mr.group
}

func (mr MergeRequest) ProjectID() int {
	return mr.projectID
}

func (mr MergeRequest) ProjectName() string {
	return mr.projectName
}

func (mr MergeRequest) ProjectWebURL() string {
	return mr.projectWebURL
}

func (mr MergeRequest) MergeReqID() int {
	return mr.mergeReqID
}

func (mr MergeRequest) MergeReqURL() string {
	return mr.mergeReqURL
}

func (mr MergeRequest) MergeReqTitle() string {
	return mr.mergeReqTitle
}

func (mr MergeRequest) WorkInProgress() bool {
	return mr.workInProgress
}

// Strip the project from the namespace path to determine the group with path
// namespace and group are the same, the web ui uses group, whilst the api uses namespace
func groupPath(namespaceWithPath string) (string, error) {
	if !strings.Contains(namespaceWithPath, "/") {
		return namespaceWithPath, errors.New("path does not contain /.")
	}
	lastSlash := strings.LastIndex(namespaceWithPath, "/")
	return namespaceWithPath[:lastSlash], nil
}

// Return details of a MergeRequest.
func (mr MergeRequest) getMR(gc GitlabWrapper) (error, *gitlab.MergeRequest) {
	options := &gitlab.GetMergeRequestsOptions{}
	result, response, err := gc.GetMergeRequest(mr.projectID, mr.mergeReqID, options)
	promGitlabReqs.WithLabelValues("merge_requests", "get", mr.group).Inc()

	if err != nil {
		return fmt.Errorf("failed to get mr: %s, http_code: %d", err, response.StatusCode), nil
	}
	return nil, result
}

// Gets list of suggested approvers from merge request - this should match the list found in the CODEOWNERS file.
func (mr MergeRequest) getMRApprovers(gc GitlabWrapper) ([]*gitlab.BasicUser, int, error) {
	result, response, err := gc.GetConfiguration(mr.projectID, mr.mergeReqID)
	promGitlabReqs.WithLabelValues("merge_requests", "get", mr.group).Inc()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get approvers: %s, http_code: %d", err, response.StatusCode)
	}

	if len(result.SuggestedApprovers) == 0 {
		promIgnoreActions.WithLabelValues("no_suggested_approvers", mr.group).Inc()
		return nil, result.ApprovalsRequired, fmt.Errorf("no suggested approvers.")
	}

	if result.ApprovalsRequired == 0 {
		promIgnoreActions.WithLabelValues("approvals_required_zero", mr.Group()).Inc()
		return nil, result.ApprovalsRequired, fmt.Errorf("approvals required is zero, will not assign reviewer.")
	}

	return result.SuggestedApprovers, result.ApprovalsRequired, nil
}

// Update MergeRequest assigning Reviewers
func (mr MergeRequest) setMRReviwer(gc GitlabWrapper, reviewers []*gitlab.BasicUser) error {
	var reviewerIDs []int
	for _, reviewer := range reviewers {
		reviewerIDs = append(reviewerIDs, reviewer.ID)
	}

	options := &gitlab.UpdateMergeRequestOptions{
		ReviewerIDs: &reviewerIDs,
	}
	_, response, err := gc.UpdateMergeRequest(mr.projectID, mr.mergeReqID, options)
	promGitlabReqs.WithLabelValues("merge_requests", "patch", mr.group).Inc()
	if err != nil {
		return fmt.Errorf("failed to assign reviewer on project: %s, http_code: %d", err, response.StatusCode)
	}
	return nil
}

// Update MergeRequest unassinging reviewers
func (mr MergeRequest) unsetMRReviwer(gc GitlabWrapper) error {
	reviewerIDs := []int{0}
	options := &gitlab.UpdateMergeRequestOptions{
		ReviewerIDs: &reviewerIDs,
	}
	_, response, err := gc.UpdateMergeRequest(mr.projectID, mr.mergeReqID, options)
	promGitlabReqs.WithLabelValues("merge_requests", "patch", mr.group).Inc()
	if err != nil {
		return fmt.Errorf("failed to unassign reviewer on project: %s, http_code: %d", err, response.StatusCode)
	}
	return nil
}
