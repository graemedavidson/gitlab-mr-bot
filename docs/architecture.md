# Architecture

Document describes technical detail of the app implementation where certain decisions require extra annotation. Not
required for general usage.

## Slack API

[Slack: api](https://api.slack.com/)

App connects to Slack through an application token created following [setup slack document](./setup-slack.md#add-app).
Token limited to requesting channel users including user details (including email and status) and posting message to
channel.

Utilises the [Slack SDK](https://github.com/slack-go/slack) for requests.

## GitLab API

[GitLab: api](https://docs.gitlab.com/ee/api/)

App connects to a GitLab instance through the use of the [GitLab Bot User](./add-gitlab-bot-user.md) with limited
permissions to work with Merge Requests.

Utilises the [GitLab SDK](https://github.com/xanzy/go-gitlab) for requests.

### Bot user identity

The app requires a [GitLab user](./add-gitlab-bot-user.md) to make changes to the MR.

The bot requests its own user identity when starting the app using the token set in the deployment.

Function: `getBotUserIdentity`

## Users

App designed to work inside a single domain (email, ldap, etc..) with shared users. Therefore the mappings from slack
user to gitlab user expected to match by an external mechanism.

The internal system relies on this shared user identifier being a unique username:

- GitLab username: `user.name`
- Slack username: `user.name`

Future work includes managing a [database use of mappings](https://github.com/graemedavidson/gitlab-mr-webhook/issues/40)
allowing for different use cases.

### Retrieve slack user meta data

User meta data pulled via the slack API.

Request workflow:

1. Requests all users in the channel specified in the configuration file. Returns a list of Slack user ids.
2. Requests required user meta data for each returned user ID in previous step.

## Logging

Utilises the [Logrus Library](https://github.com/sirupsen/logrus).

Current logging focused on tracking errors and debug statements for runtime workflow.

[Future work](https://github.com/graemedavidson/gitlab-mr-webhook/issues/11) required to standerdise logging output.

## Scheduler

Scheduler sets up go channels and runs loop for checking state of channels and handling responses from workers.
Essentially imitating a queue.

Startup [(`main.go`)](../main.go):

- Scheduler created
- Workers created and registered to scheduler
    - Number of workers determined by the number of CPU cores available
- Runs scheduler in [goroutine](https://www.golang-book.com/books/intro/10)

Scheduler [(`scheduler.go`)](../scheduler.go):

- Create required channels:
    - `requests`: MergeRequest payloads ready for processing by worker pool
    - `responses`: responses from workers processing
    - `status`: snapshot of the current status of workers (working/idle)
- Create workers in [goroutines](https://www.golang-book.com/books/intro/10)
    - `go worker.Run(...)`
    - passes reference to all channels
- Run `messagePump` loop checking for worker response and statuses
    - Currently schedulers `handleResponse` function is empty as no action required

### Workers

[Worker Source](../worker.go)

- Each worker runs in its own go routine
    - number determined by scheduler (see preceeding section).
- Runs in continuous loop checking go channel for new MRs
- MR sent to channel, worker recieves the MR payload and processes payload

## Configuration file

see [deployment: configuration file](./deployment.md#configuration-file)

## Web server

The app hosts a basic web server implementation use the `net/http` package, hosted pages:

| Endpoint      | description
| ---           | ---
| `/webhook`    | endpoint called by gitlab when setting up the webhook
| `/metrics`    | prometheus metrics
| `/health`     | health check endpoint including checking version
| `/cache`      | UI for managing user status cache
| `/static`     | Static assets for UI

## Telemetry

Telemetry uses prometheus metrics for collecting metrics. [Metrics.](./telemetry.md)

Utilises the [prometheus client](https://github.com/prometheus/client_golang) to host a metrics endpoint over the
web server at the endpoint `/metrics`.

## Selecting random users

Review [create a merge request](./creating-an-mr.md) for a high level description of how the selection process.

## User status cache

User status cache stores a users slack status used to determine availability for selection to approve an MR. When
processing an MR the app checks the status of the users cache entry, if the user exists and the cache ttl is valid it
uses the known status.

If user cache entry expired or doesn't exist the app makes an outbound call to Slack to
[retrieve the user data](#retrieve-slack-user-meta-data) and add to the cache.

## Ignore app changes to MR

When the app makes a change to an MR to remove or add a reviewer the interaction causes a new webhook call. To
circumvent this behaviour causing further changes to an MR the app ignores all modifications where the author of the
change is its own bot user. Tracked via [telemetry](./monitoring.md): `promRecursiveCalls`.
