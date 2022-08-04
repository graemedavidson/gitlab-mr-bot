# Gitlab setup and configuration

App requires a user with access to the GitLab Project and Merge Requests to make changes.

## Create Bot User

- [GitLab: creating Users](https://docs.gitlab.com/ee/user/profile/account/create_accounts.html)

### Create Token

Token passed into the App as a environment variable and used for API access.

**Note:** Token security is bespoke to deployment with usual security expectations of generating/storing in a secure
location and rotating.

- [GitLab: personal access tokens](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)

## Adding to project

Requires `developer` access to make changes to a project MR.

- [GitLab: permissions](https://docs.gitlab.com/ee/user/permissions.html).
- [GitLab: members of a project](https://docs.gitlab.com/ee/user/project/members/)
