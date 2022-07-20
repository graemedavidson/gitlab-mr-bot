# Changelog
https://keepachangelog.com/en/1.0.0/

## [v0.10.0] - 03/08/2022
### Changed
- Repoistory migrated from private to public with open source license
- Reset versioning
- Docs written for general use

---

# Changes made under old repository

## [v0.9.0] - 02/06/2022
### Changed
- Logging outputs to be more concise and remove duplicates.
- `gitlab_mr_wh_slack_errors` -> `gitlab_mr_wh_slack_msgs_errors` to better reflect it being messaging errors only.

### Added
- slack api prom metrics: `gitlab_mr_wh_slack_api_reqs` and `gitlab_mr_wh_slack_api_errors` for recording slack api
  requests and errors.
- cache prom metrics: `gitlab_mr_wh_cache_delete`, `gitlab_mr_wh_cache_clear` and `gitlab_mr_wh_cache_admin` for
  recording cache actions.
- user statuses for ttl settings through config file.
- Cache admin user interface for administrating the cache entries
- github action: go unit tests
- example grafana dashboard to local env
- worker prom metrics (`gitlab_mr_wh_workers`, `gitlab_mr_wh_workers_working`)

## [v0.8.3] - 20/05/2022
### Fixed
- response header from `204` (No Content) to `202` (Accepted) which allows for a body response and better reflects the
  status of the merge request which has been accepted into the processing queue. Processing and status are then handled
  asynchronously and will not be passed onto the webhook request.

## [v0.8.2] - 18/05/2022
### Fixed
- missing ids from cache to be based on passing usernames and
  checking for missing slack channel ids
  - also added clause so that if cache empty return all
- cached status: `Out Sick` -> `Out sick`

### Updated
- process mr to handle no available approvers after checking
  slack statuses

### Removed
- regressed error in worker:getSlackUserIDs which was still
  checking cache with slack channel ids. Now returns all ids found to be
  added to the cache.
  - added to do into the code regarding potential for large volume of
    responses which might require pagniation.
  - removed cached mock from unit tests

### Added
- project name to slack message

## [v0.8.1] - 17/05/2022
### Updated
- development and build env to go 1.18
- all go mods to latest versions

## [v0.8.0] - 25/04/2022
### Changed
- random approvers logic to be part of worker and not as part of the merge request.

### Added
- slack user status for determining if users who are sick or on vacation should be ignored as a suggested approver.
    - slack api setup with token
        - documentation included for generating application and token and adding to slack and channels

## [v0.7.0] - 18/12/2021
### Added
- Scheduler for managing multiple workers.
- Simple responses from workers handling MR, this requires more work to improve overall logging and metrics.
- Statuses of workers are shown when a worker completes a task.
- Automaxprocs for determining number of workers created within k8s

## [v0.6.2] - 2/12/2021
### Added
- first pass of unit tests

### Changed
- files and code structure to better separate areas of concern. Changing and moving structures as required.
- logging setup to have a default set of labels, this likely can be further improved

## [v0.6.1] - 2/12/2021
### Changed
- Use the full path for slack channel mapping. Allows for single repository mappings.

## [v0.6.0] - 2/12/2021
### Removed
- removed tests failing metric as did not best describe actions of bot

### Updated
- All events pushing to tests failing metric to use ignored actions metric as this better describes the bot action and
  simplifies.
- grafana dashboard example

## [v0.5.0] - 1/12/2021
### Fixed
- updated prometheus metrics for gitlab requests to use 'patch' when making changes

### Added
- removing the active reviewers when setting the project to a WIP.
- promRemoveReviewer metric for counting times a reviewer is removed.
    - Added graph to grafana dashboard

## [v0.4.0] - 18/11/2021
### Updated
- readme to include new information on configuration file
- grafana dashboard example to include
    - ignored actions
    - groups
    - slack message errors

### Changed
- failing checks to no longer return an error as this was an abuse of http error codes.
- setting WIP/Draft actions to be included in the new ignored actions metric
- Added group label to ignored actions, slack messages and events metrics

## [v0.3.1] - 18/11/2021
### Added
- Check event for the action, if it is one of the following it will be ignored as post this event a reviewer being set
  would be unnecessary if one has not already assigned. This can happen when MRs were created before the review bot was
  setup or if the review bot was unavailable.
    - `approved` - An approve event means the MR has been approved by a team member and therefore does not need further
      approval.
    - `merge` - The merge request has been approved and therefore does not require reviewing.
- Prometheus metric for ignored actions including tags for action and namespace.

## [v0.3.0] - 18/11/2021
### Changed
- Slack message to include Merge Request Title instead of raw URL.

## [v0.2.0] - 17/11/2021
### Added
- setting the slack channel via the namespace (gitlab group) a merge request belongs to.
- Pulls in yaml configuration file to set the namespace (gitlab group) to slack channel mappings
- namespace (gitlab group) to logging and prometheus metrics

### Changes
- refactors passing of vars to be via structs

## [v0.1.1]
### Changed
- Refactored files to better separate by functionality

## [v0.1.0]
### Added
- Proof of concept codebase for running a webhook in Gitlab which when called assigns a random reviewer to a merge
  request picked for the suggested approvers set in the CODEOWNERS file.
- Local development environment
