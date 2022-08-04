package main

import (
	"github.com/slack-go/slack"
	// "context"
)

type SlackWrapper interface {
	PostWebhook(msg *slack.WebhookMessage) error
	GetUsersInfo(users ...string) (*[]slack.User, error)
	GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error)
	webhookURLSet() bool
}

type Slack struct {
	client *slack.Client
	wh_url string
}

func (s *Slack) PostWebhook(msg *slack.WebhookMessage) error {
	// Webhook sends without client context as auth in message url, will likely update this.
	return slack.PostWebhook(s.wh_url, msg)
}

func (s *Slack) GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error) {
	return s.client.GetUsersInConversation(params)
}

func (s *Slack) GetUsersInfo(users ...string) (*[]slack.User, error) {
	return s.client.GetUsersInfo(users...)
}

func (s *Slack) webhookURLSet() bool {
	return s.wh_url != ""
}

func newSlackClient(token string, wh_url string) *Slack {
	api := slack.New(token)
	return &Slack{client: api, wh_url: wh_url}
}
