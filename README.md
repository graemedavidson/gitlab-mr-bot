# Gitlab MR Bot

The Gitlab MR Bot provides a webhook service which assigns a random reviewer to a Merge Request. After assigning a user
from the suggested approver list such as from the CODEOWNERS file it can send a notification to a configured slack
channel.

The bot is currently developed against [GitLab Premium](https://about.gitlab.com/pricing/) edition which allows for
adding reviewers and using CODEOWNERS files.

## State of development

Fair warning, codebase maintained by a infrastructure engineer who was learning GO at the time and certainly would not
call themselves a software engineer. Coded as a side project and a bit of fun. Happy to take on any
suggestions/improvements, please raise an issue or see the [contributing](#contributing) section to help out.

## Gitlab should do this

You might be thinking that Gitlab should offer this within their integrations, others agree, have a search through
their gitlab issues and find it mentioned a few times. This project is starting to extend feature requests beyond what
they might offer but who knows, this might all be academic some point soon.

## Known bugs / Tasks remaining

Please review [issues](https://github.com/graemedavidson/gitlab-mr-bot/issues) in this repository for remaining work
identified. Please add an issue should you discover a bug or would like to request a feature.

## Running app

The App requires some passed in configuration via the command line or environment variable. See
[deployments documentation](./docs/deployment.md) for more information on this subject.

## Contributing

To contribute please first create an [issue](https://github.com/graemedavidson/gitlab-mr-bot/issues) to use as a ref in
commits. Review preceeding section "Known bugs / Tasks remaining" to ensure no existing issue describing bug/feature.

When ready create an MR with the following changes included:

- Changelog including a description from user perspective on what the change does
- Concise commits including issue ref and technical description of why a change is happening
- Tests for new features and old ones still passing
- Documentation

### Local development

**Note:** currently not working due the limited capability of a free gitlab version.

Basic configuration for running the required components through docker.

## Documentation

- [Usage](./docs/usage.md)

## Credits / References

- [GitLab: api](https://docs.gitlab.com/ee/api/api_resources.html)
    - [Merge requests](https://docs.gitlab.com/ee/api/merge_requests.html)
    - [Merge request approvals](https://docs.gitlab.com/ee/api/merge_request_approvals.html)
    - [Webhook server example](https://github.com/xanzy/go-gitlab/blob/master/examples/webhook.go)
- [Go: GitLab SDK](https://github.com/xanzy/go-gitlab)
- Slack
    - [Go: Slack SDK](https://github.com/slack-go/slack)
    - [Slack: build app](https://api.slack.com/start/building)
- [Go: prometheus instrumentation](https://prometheus.io/docs/guides/go-application/)
- [Go: logging](https://github.com/sirupsen/logrus)
