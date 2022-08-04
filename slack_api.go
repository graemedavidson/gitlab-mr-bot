package main

import (
	"github.com/slack-go/slack"
	// "context"
)

type SlackWrapper interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	GetUsersInfo(users ...string) (*[]slack.User, error)
	GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error)
}

type Slack struct {
	client *slack.Client
}

func (s *Slack) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	return s.client.PostMessage(channelID, options...)
}

func (s *Slack) GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error) {
	return s.client.GetUsersInConversation(params)
}

func (s *Slack) GetUsersInfo(users ...string) (*[]slack.User, error) {
	return s.client.GetUsersInfo(users...)
}

func newSlackClient(token string) *Slack {
	api := slack.New(token)
	return &Slack{client: api}
}
