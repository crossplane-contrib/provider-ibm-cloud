# provider-ibm-cloud

## Overview

`provider-ibm-cloud` is the Crossplane infrastructure provider for [IBM Cloud](https://cloud.ibm.com). The provider that is built from the source
code in this repository can be installed into a Crossplane control plane and
adds the following new functionality:

* Custom Resource Definitions (CRDs) that model IBM Cloud infrastructure and services
  (e.g. [Resource Controller](https://cloud.ibm.com/apidocs/resource-controller/resource-controller), 
  [IAM](https://cloud.ibm.com/apidocs/iam-access-groups/), etc.)
* Controllers to provision these resources in IBM Cloud based on the users desired
  state captured in CRDs they create.
* Implementations of Crossplane's portable resource abstractions, enabling IBM Cloud
  resources to fulfill a user's general need for cloud services.

## Getting Started and Documentation

For getting started guides, installation, deployment, and administration, see
our [Documentation](https://crossplane.io/docs/latest).

To use the IBM Cloud Provider you'll need an IBM Cloud account. You may sign up 
for a [free IBM Cloud account](https://cloud.ibm.com/registration). You'll need also an IBM 
Cloud CLI which needs to be [installed and configured](https://cloud.ibm.com/docs/cli).
Make sure you have a Kubernetes cluster and [installed Crossplane](https://crossplane.io/docs/v0.14/getting-started/install-configure.html), then:

### Install IBM Cloud Provider

```shell
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1alpha1
kind: Provider
metadata:
  name: provider-ibm-cloud
spec:
  package: "crossplane/provider-ibm-cloud:alpha"
EOF
```

### Generate IBM Cloud API Key

Using an IBM Cloud account with permissions to manage IBM Cloud Services:

```shell
if [[ -z "${IBMCLOUD_API_KEY}" ]]; then
  echo "*** Generating new APIKey"
  IBMCLOUD_API_KEY=$(ibmcloud iam api-key-create provider-ibm-cloud-key -d "Key for Crossplane Provider IBM Cloud" | grep "API Key" | awk '{ print $3 }')
fi
```

### Create a Provider Secret

```shell
kubectl create secret generic provider-ibm-cloud-secret --from-literal=credentials=${IBMCLOUD_API_KEY} -n crossplane-system
```

### Configure the Provider

We will create the following ProviderConfig object to configure credentials for IBM Cloud Provider:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: ibm-cloud.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: ibm-cloud
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: provider-ibm-cloud-secret
      key: credentials
  region: us-south
EOF
```
### Next Steps

Now that you have the IBM Cloud provider configured, you can [provision infrastructure](https://crossplane.io/docs/v0.14/getting-started/provision-infrastructure.html). See [examples](examples) for the IBM Cloud Provider.

## Contributing

provider-ibm-cloud is a community driven project and we welcome contributions. See the
Crossplane [Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/crossplane/provider-ibm-cloud/issues).

## Contact

Please use the following to reach members of the community:

* Slack: Join our [slack channel](https://slack.crossplane.io)
* Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
* Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
* Email: [info@crossplane.io](mailto:info@crossplane.io)

## Roadmap

provider-ibm-cloud goals and milestones will be tracked in the Crossplane
repository. More information can be found in
[ROADMAP.md](https://github.com/crossplane/crossplane/blob/master/ROADMAP.md).

## Governance and Owners

provider-ibm-cloud is run according to the same
[Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md)
and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md)
structure as the core Crossplane project.

## Code of Conduct

provider-ibm-cloud adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

provider-ibm-cloud is under the Apache 2.0 license.

[![FOSSA
Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fcrossplane-contrib/%2Fprovider-ibm-cloud.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fcrossplane-contrib%2Fprovider-ibm-cloud?ref=badge_large)
