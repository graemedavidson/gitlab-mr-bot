# Debugging issues

## Check the webhook is working

### Check configured in repository

Requirements:

- maintainer or greater permissions - [GitLab: permissions](https://docs.gitlab.com/ee/user/permissions.html)

**Note:** `edit` option reveals the `Secret Token` used as part of the authentication to the Gitlab MR WH service, refer
to best security practises to keep safe.

Navigate to project [webhooks](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html).

`Settings > Webhooks`

Options:

- `edit` option to change the webhook configuration and review webhook call history including payload and response.
- `test` runs test configuration against the webhook

### Check project correctly setup

Ensure all elements configured at the [project level](./setup-gitlab-project.md).

### Review app responses

**Note:** Minimal design in current HTTP responses, review [future development](./development-roadmap.md) for expected
changed.

#### Server Errors

http response code: `5xx`

Error occurred when handling the merge request payload, review response body for error details.

#### Successful accepted

http response code: `202`

MR event handled without incident. Potential responses:

> Ignoring request initiated through this service

When the service sets a reviewer it itself modifies the Merge Request therefore setting off the webhook again, this
response indicates that occurring. The app ignore these requests.

> Ignoring approved/merge action

GitLab MR payload `action` equals `approved` or `merge` indicating that assigning a reviewer no longer required.

This scenario occurs in situations where MR created before the app configured or during a downtime.

> Ignoring as merge request cannot be merged

Merge request has conflicts which require resolving before assigning for review.

> Successfully added merge request to processing queue

All checks passed, assign a reviewer.
