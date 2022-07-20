# Local Development Environment

The local development is comprised of all the components required to run the merge request reviewer in production. It is
made up of the following services:

* Gitlab CE - A local vesion of Gitlab which is useful for testing a webhook lifecycle. This might be replaced in the
  future with a mocked system as running gitlab is quite intensive and slow.
* Monitoring
    * [Grafana](https://grafana.com/)
    * [Prometheus](https://prometheus.io/)
    * [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)

## Gitlab CE

The following [documentation](https://docs.gitlab.com/ee/install/docker.html#pre-configure-docker-container) provides
more information on running Gitlab locally.

Once Gitlab is running run the following command to grab the password.

```bash
docker exec -it <GITLAB_CONTAINER> grep 'Password:' /etc/gitlab/initial_root_password
```

Gitlab must then be configured to [allow local webhooks](https://docs.gitlab.com/ee/security/webhooks.html)
