# Setup GitLab Project

## Bot User Access

To access and assign MRs the App requires a user with developer permissions at the group or project level. See
[Gitlab setup and configuration](./add-gitlab-bot-user.md) for more information.

## Setup Webhook

The following example makes use of terraform to deploy the webhook config to all repos under a group. It requires a
gitlab token with `owner` access to the group. It makes use of a vault instance for storing the secret token.

*Note:* the web hook token is accessible from the Gitlab UI for `owner` level users.

```terraform
data "vault_generic_secret" "webhook_config" {
  path = "infra/tf/gitlab/mergerequest-webhook/${var.tier}"
}

data "gitlab_group" "group" {
  full_path = "target-group"
}

# review pagination
data "gitlab_projects" "projects" {
  group_id          = data.gitlab_group.group.id
  order_by          = "name"
  include_subgroups = true
  simple            = true
  with_shared       = false

  # data query config
  per_page            = 100
  max_queryable_pages = 20
}

locals {
    exclude_projects = [
        "target-group/exclude-this-project",
    ]
}

resource "gitlab_project_hook" "projects" {
  for_each = { for project in data.gitlab_projects.projects.projects : project.path_with_namespace => project.id if !contains(local.exclude_projects, project.path_with_namespace) }

  project = each.key

  url   = "http://gitlab.example:8080/webhook"
  token = data.vault_generic_secret.webhook_config.data["webhook_secret"]

  # configure only for merge request events
  merge_requests_events = true
}
```

## CODEOWNERS

The recommended method of adding a list suggested approvers is through the [CODEOWNERS file](https://docs.gitlab.com/ee/user/project/code_owners.html).
Approval from code owners can be enables as part of the protected branches section:

`Settings > Repository` and expand the `Protected branches`. Add or edit an existing protected branch checking the
'Require approval from code owners' option.

## Required Approval

Default required approvers set to `0`, therefore no approvers set via the bot by default. Update at the repository
level:

`Settings > General > Merge request approvals` section.

[GitLab Docs](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/settings.html)

## Turn off author approvals

Deny author approvals as otherwise the App can randomly assign the author to their own MR. [Turn off](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/settings.html#prevent-approval-by-author).
