package main

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

type GitlabWrapper interface {
	GetMergeRequest(pid interface{}, mergeRequest int, opt *gitlab.GetMergeRequestsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error)
	CurrentUser(options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error)
	GetConfiguration(pid interface{}, mr int, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequestApprovals, *gitlab.Response, error)
	UpdateMergeRequest(pid interface{}, mergeRequest int, opt *gitlab.UpdateMergeRequestOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error)
}

type Gitlab struct {
	client *gitlab.Client
}

func (g *Gitlab) GetMergeRequest(pid interface{}, mergeRequest int, opt *gitlab.GetMergeRequestsOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
	return g.client.MergeRequests.GetMergeRequest(pid, mergeRequest, opt)
}

func (g *Gitlab) CurrentUser(options ...gitlab.RequestOptionFunc) (*gitlab.User, *gitlab.Response, error) {
	return g.client.Users.CurrentUser()
}

func (g *Gitlab) GetConfiguration(pid interface{}, mr int, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequestApprovals, *gitlab.Response, error) {
	return g.client.MergeRequestApprovals.GetConfiguration(pid, mr)
}

func (g *Gitlab) UpdateMergeRequest(pid interface{}, mergeRequest int, opt *gitlab.UpdateMergeRequestOptions, options ...gitlab.RequestOptionFunc) (*gitlab.MergeRequest, *gitlab.Response, error) {
	return g.client.MergeRequests.UpdateMergeRequest(pid, mergeRequest, opt)
}

func newGitlabClient(host string, token string) (*Gitlab, error) {
	c, err := gitlab.NewClient(token, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", host)))
	if err != nil {
		return nil, err
	}
	return &Gitlab{client: c}, nil
}
