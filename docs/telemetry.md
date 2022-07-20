# Telemetry

Metrics exposed via `metrics` endpoint utilising the [prometheus client](https://github.com/prometheus/client_golang).

[Prometheus: metric types](https://prometheus.io/docs/concepts/metric_types/)

[Source code](../prometheus_metrics.go)

| Name                                      | Type      | Labels                        | Description
| ---                                       | ---       | ---                           | ---
| `gitlab_mr_wh_events`                     | Counter   | `type`, `group`               | The total number of gitlab events.
| `gitlab_mr_wh_processed_mrs`              | Counter   | `group`                       | The total number of gitlab merge requests that require assigning reviewers.
| `gitlab_mr_wh_review_removed`             | Counter   | `group`                       | The total number of times a merge request is set to WIP and then a reviewer removed.
| `gitlab_mr_wh_recursive_calls`            | Counter   | `group`                       | When this service calls itself after updating a merge request activing the webhooks.
| `gitlab_mr_wh_errors`                     | Counter   | `error`                       | The total number of errors encountered handling events.
| `gitlab_mr_wh_ignore_actions`             | Counter   | `action`, `group`             | The total number of ignored actions (merge, approve) handling events.
| `gitlab_mr_wh_slack_api_reqs`             | Counter   | `request`                     | The total number of slack api requests
| `gitlab_mr_wh_slack_api_errors`           | Counter   | `request`, `error`            | The total number of slack api request errors
| `gitlab_mr_wh_slack_msgs`                 | Counter   | `group`, `channel`            | The total number of slack messages sent.
| `gitlab_mr_wh_slack_msgs_errors`          | Counter   | `error`, `group`, `channel`   | Errors encountered when attempting to send slack messages
| `gitlab_mr_wh_gitlab_reqs`                | Counter   | `request`, `method`, `group`  | The total number of gitlab requests made.
| `gitlab_mr_wh_cache_read`                 | Counter   | `response`, `reason`          | Cache reads with hit/miss labels with a reason for miss.
| `gitlab_mr_wh_cache_updates`              | Counter   |                               | Cache updates.
| `gitlab_mr_wh_cache_delete`               | Counter   |                               | Cache delete.
| `gitlab_mr_wh_cache_clear`                | Counter   |                               | Cache entry cleared.
| `gitlab_mr_wh_cache_admin`                | Counter   |                               | Cache Admin page accessed.
| `gitlab_mr_wh_no_matching_slack_user`     | Counter   | `group`                       | A gitlab user in the codeowners does not have a matching entry in slack.
| `gitlab_mr_wh_slack_status_unavailable`   | Counter   | `reason`, `group`             | Slack status of user means they are unavailable as an approver.
| `gitlab_mr_wh_workers`                    | Counter   |                               | Number of workers created.
| `gitlab_mr_wh_workers_working`            | Guage     |                               | Number of workers working.
