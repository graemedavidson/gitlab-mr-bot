# Slack Statuses

The app checks the current status of a user via Slack and based on the status retains or removes the user from the
suggested approvers pool before random selection.

## Setting status

[Slack: set your Slack status and availability](https://slack.com/intl/en-gb/help/articles/201864558-Set-your-Slack-status-and-availability)

### Statuses

The following set statuses removes a user from the selected pool.

| Status       | Cache Time (ttl)
| ---          | ---
| `Holiday`    | 8 hours
| `Vactioning` | 8 hours
| `Out sick`   | 8 hours

## Caching

App uses a local in memory cache to store a slack status against a user. Cached response returned until the cache ttl
expires. Upon expiry app requests updated slack and updates cache.

## Setting in configuration file

Custom statuses set in the [deployment configuration file.](./deployment.md#configuration-file)
