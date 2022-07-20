# Application deployment

## Requirements

- Gitlab Bot User Token: App requires a GitLab user token to access the GitLab API to pull in Merge Request data.
    - [GitLab Bot User](./add-gitlab-bot-user.md)

- Slack Bot User OAuth Token: App requires a slack App token to access slack API to pull in user data and write messages.
    - [Slack](./setup-slack.md#oauth-token)

## Deployment

An App deployment consists of running an app with an accessible URL from Gitlab. Proceeding text and code is an Example
deployment using Terraform and Kubernetes. Source code in the [examples](./examples) directory.

**Note:** the following example exposes a web server with only a token for security. Consider other security factors in
a production deployment such as a [Nginx reverse proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/).

### Environment variables

| Environment Variable          | Default | Description
| ---                           | ---     | ---
| `GITLAB_TOKEN`                |         | [Gitlab bot user token](#gitlab-bot-user-token)
| `GITLAB_MR_WH_LOG_LEVEL`      | `Warn`  | Logging [level](https://github.com/sirupsen/logrus#level-logging) of app
| `GITLAB_URL`                  |         | URL of gitlab (example: gitlab.local)
| `GITLAB_MR_WH_WEBHOOK_SECRET` |         | Secret token passed with MR payload set when adding webhook in [project setup](./setup-gitlab-project.md#setup-webhook)
| `GITLAB_MR_WH_SLACK_TOKEN`    |         | Slack OAuth token used for API calls to Slack Workspace
| `GITLAB_MR_WH_LISTEN_PORT`    | `8080`  | Port for app server

### Configuration file

The MR Bot as part of the deployment includes a configuration which stores the slack channel name and ID matched to
Gitlab Groups full path. The group determines the messaged channel when processing an MR.

Configuration expected in the following directory relative to the app deployment:

```
./config/config.yaml
```

```yaml
---
group_channels:
  gitlab:
    slack_channel: "#gitlab-notifications"
    slack_channel_id: "1A1A1A1A1"
  mrs:
    slack_channel: "#mrs"
    slack_channel_id: "2A2A2A2A2"

user_statuses:
  "": 1
  "out sick": 8
  "holiday": 8
  "vacationing": 8
```

Subgroup configuration:

```yaml
---
group_channels:
  gitlab:
    slack_channel: "#gitlab-notifications"
    slack_channel_id: "1A1A1A1A1"
  gitlab/mrs:
    slack_channel: "#gitlab-mrs"
    slack_channel_id: "3A3A3A3A3"

user_statuses:
  "": 1
  "out sick": 8
  "holiday": 8
  "vacationing": 8
```

### Channel ID

A Slack Channel ID is available through the UI by expanding the `Get channel details` button when on a channel.

[![get slack channel id 1](./images/slack-get-channel-id-1.png)](./images/slack-get-channel-id-1.png)

Then scrolling to the bottom where it displays the ID.

[![get slack channel id 2](./images/slack-get-channel-id-2.png)](./images/slack-get-channel-id-2.png)
