# Brigade Metrics: One-Step Monitoring for Brigade 2

Brigade Metrics adds monitoring capabilities to a Brigade 2 installation. It
utilizes Brigade APIs to export time series metrics to Prometheus and makes
visualizations of those metrics available through a Grafana dashboard.

Brigade 2 itself is currently in an _beta_ state and remains under active
development, as such, the same is true for this add-on component.

## Getting Started

Comprehensive documentation will become available in conjunction with a future
release. In the meantime, here is a little to get you started.

### Prerequisites

Since Brigade Metrics aggregates and exposes metrics for a running Brigade 2
installation, an operation Brigade 2 beta installation is a prerequisite. Please
refer to
[Brigade 2's own getting started documentation](https://github.com/brigadecore/brigade/tree/v2).

Once Brigade 2 is up and running, create a service account, and give it `READ`
permissions:

```console
$ brig sa create -i brigade-metrics -d brigade-metrics
$ brig role grant READER --service-account brigade-metrics
```

Save the service account token somewhere safe.

### Installing Brigade Metrics

Since this add-on is still very much a prototype, we're not currently publishing
a Helm chart anywhere. You will need to clone this repository to obtain the
chart and install.

Once the repository is cloned, open the `values.yaml` file, and paste the
service account token into the `exporter.brigade.apiToken` field.

There are two methods of authentication you can choose from for logging into Grafana. 
1. Option to use Grafana's built in user management system. The username and password for the admin account are specified in the `grafana.auth` fields, and the admin can handle user management using the Grafana UI.
2. Option to use an nginx reverse proxy and a shared username/password to access Grafana in anonymous mode.

For option 1, set `grafana.auth.proxy` to false in `values.yaml`, and true for option 2.

In addition, you have the option to enable tls or ingress for grafana, and both options can be configured in `values.yaml`.

Save the file, and run `make hack` from the project's root directory.

Once all three pods of the project are up and running, run the following command to expose the Grafana frontend:

```console
$ kubectl port-forward service/brigade-metrics-grafana 3000:<80 (tls disabled), 443 (tls enabled)> -n brigade-metrics
```

Enter your supplied credentials. You can now access the Grafana dashboard!

## Contributing

The Brigade project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature open an issue
[here](https://github.com/brigadecore/brigade/issues)
