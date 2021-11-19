# Brigade Metrics: Monitoring for Brigade 2

![build](https://badgr.brigade2.io/v1/github/checks/brigadecore/brigade-metrics/badge.svg?appID=99005)
[![codecov](https://codecov.io/gh/brigadecore/brigade-metrics/branch/main/graph/badge.svg?token=ER6NYB0V9K)](https://codecov.io/gh/brigadecore/brigade-metrics)
[![Go Report Card](https://goreportcard.com/badge/github.com/brigadecore/brigade-metrics)](https://goreportcard.com/report/github.com/brigadecore/brigade-metrics)
[![slack](https://img.shields.io/badge/slack-brigade-brightgreen.svg?logo=slack)](https://kubernetes.slack.com/messages/C87MF1RFD)

<img width="100" align="left" src="logo.png">

Brigade Metrics adds monitoring capabilities to a Brigade 2 installation. It
utilizes Brigade APIs to export time series metrics to Prometheus and makes
visualizations of those metrics available through a Grafana dashboard.

<br clear="left"/>

## Installation

Prerequisites:

* A Kubernetes cluster:
    * For which you have the `admin` cluster role
    * That is already running Brigade 2
    * Capable of provisioning a _public IP address_ for a service of type
      `LoadBalancer`. (This means you won't have much luck running the gateway
      locally in the likes of kind or minikube unless you're able and willing to
      mess with port forwarding settings on your router, which we won't be
      covering here.)

* `kubectl`, `helm` (commands below require Helm 3.7.0+), and `brig` (the
  Brigade 2 CLI)

### 1. Create a Service Account

__Note:__ To proceed beyond this point, you'll need to be logged into Brigade 2
as the "root" user (not recommended) or (preferably) as a user with the `ADMIN`
role. Further discussion of this is beyond the scope of this documentation.
Please refer to Brigade's own documentation.

Using Brigade 2's `brig` CLI, create a service account:

```console
$ brig service-account create \
    --id brigade-metrics \
    --description brigade-metrics
```

Make note of the __token__ returned. This value will be used in another step.
_It is your only opportunity to access this value, as Brigade does not save it._

Authorize this service account with read-only access to Brigade:

```console
$ brig role grant READER \
    --service-account brigade-metrics
```

### 2. Installing Brigade Metrics

For now, we're using the [GitHub Container Registry](https://ghcr.io) (which is
an [OCI registry](https://helm.sh/docs/topics/registries/)) to host our Helm
chart. Helm 3.7 has _experimental_ support for OCI registries. In the event that
the Helm 3.7 dependency proves troublesome for users, or in the event that this
experimental feature goes away, or isn't working like we'd hope, we will revisit
this choice before going GA.

First, be sure you are using
[Helm 3.7.0](https://github.com/helm/helm/releases/tag/v3.7.0) or greater and
enable experimental OCI support:

```console
$ export HELM_EXPERIMENTAL_OCI=1
```

Use the following command to extract the full set of configuration options from
the chart. Here we're storing a copy at `~/brigade-metrics-values.yaml`:

```console
$ helm inspect values oci://ghcr.io/brigadecore/brigade-metrics \
    --version v0.3.0 > ~/brigade-metrics-values.yaml
```

Edit the configuration (`~/brigade-metrics-values.yaml` in this example). At
minimum, you will need to make the following changes:

* Set the value of `exporter.brigade.apiAddress` to the address of your Brigade 2
  API server. This should utilize the _internal_ DNS hostname by which that API
  server is reachable _within_ your Kubernetes cluster. This value is defaulted
  to `https://brigade-apiserver.brigade.svc.cluster.local`, but may need to be
  updated if you installed Brigade 2 in a different namespace.

* Set the value of `exporter.brigade.apiToken` to the service account token that
  was generated earlier.

* `grafana.host`: Set this to the host name where you'd like the dashboard
  (Grafana) to be accessible.

* Specify a username and password for the metrics dashboard by setting values
  for `grafana.auth.username` and `grafana.auth.password`.

* `grafana.service.type`: If you plan to enable ingress (advanced), you can
  leave this as its default -- `ClusterIP`. If you do not plan to enable
  ingress, you probably will want to change this value to `LoadBalancer`.

Install Brigade Metrics, referencing your edited configuration:

```console
$ helm install brigade-metrics \
    oci://ghcr.io/brigadecore/brigade-metrics \
    --version v0.3.0 \
    --create-namespace \
    --namespace brigade-metrics \
    --values ~/brigade-metrics-values.yaml \
    --wait \
    --timeout 300s
```

### 3. (RECOMMENDED) Create a DNS Entry

If you overrode defaults and set `grafana.service.type` to `LoadBalancer`, use
this command to find the gateway's public IP address:

```console
$ kubectl get svc brigade-metrics-grafana \
    --namespace brigade-metrics \
    --output jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

If you overrode defaults and enabled support for an ingress controller, you
probably know what you're doing well enough to track down the correct IP without
our help. 😉

With this public IP in hand, edit your name servers and add an `A` record
pointing your domain to the public IP.

### 4. Accessing the Dashboard

If you overrode defaults and set `grafana.service.type` to `LoadBalancer`, then
the dashboard should be accessible over HTTPS at the public IP address or DNS
hostname.

If you kept the default setting of `ClusterIP` for `grafana.service.type`, then
use port forwarding to expose the dashboard on your local network interface:

```console
$ kubectl port-forward \
    service/brigade-metrics-grafana \
    --namespace brigade-metrics \
    8443:443
```

In this case, the dashboard should be accessible at `https://localhost:8443`.
Expect to receive a cert warning.

Log in using the username and password you selected in the previous section.

## Contributing

The Brigade project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature open an issue
[here](https://github.com/brigadecore/brigade-metrics/issues)

## Code of Conduct

Participation in the Brigade project is governed by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
