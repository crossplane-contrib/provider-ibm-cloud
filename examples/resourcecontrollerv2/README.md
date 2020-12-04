# Resource Controller V2 API examples

The Resource Controller V2 API allows to provision and manage hosted services and their credentials 
from the IBM Cloud Catalog (85+ RC-compatible services).

# Importing Existing Resource Instances

Crossplane allows to [import and manage resources created outside Crossplane](https://github.com/crossplane/crossplane.github.io/blob/master/docs/master/introduction/managed-resources.md#importing-existing-resources).

For example, to import a resource instance created outside Crossplane
you may retrieve the ID and the required parameters for the resource with the
command:

```shell
ibmcloud resource service-instance <resource-instance-name>
```

The ID should be set in the `crossplane.io/external-name:` annotation and the required parameters
should be set in the `forProvider` section. This could be also scripted as follows:

First, note the name of the resource and run:

```shell
NAME=<resource-instance-name>
```

then, run the following script to get and parse the required parameters:

```shell
INFO="$(ibmcloud resource service-instance "${NAME}")"
ID=$(echo "$INFO" | grep ^ID | awk '{print $2}')
TARGET=$(echo "$INFO" | grep ^Location | awk '{print $2}')
SERVICE_NAME=$(echo "$INFO" | grep "^Service Name" | awk '{print $3}')
RG_NAME=$(echo "$INFO" | grep "^Resource Group Name" | awk '{print $4}')
RP_NAME=$(echo "$INFO" | grep "^Service Plan Name" | awk '{print $4}')
META_NAME=$(echo "$NAME" | awk '{print tolower($0)}' | tr " " - | tr "." -)
```

finally, create and apply the custom resource:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: resourcecontrollerv2.ibm-cloud.crossplane.io/v1alpha1
kind: ResourceInstance
metadata:
  name: $META_NAME
  annotations:
    crossplane.io/external-name: "$ID"
spec:
  forProvider:
    name: $NAME
    target: $TARGET
    resourceGroupName: $RG_NAME
    serviceName: $SERVICE_NAME
    resourcePlanName: $RP_NAME
  providerConfigRef:
    name: ibm-cloud
EOF
```