<ins>Read the following to the very end before you try anything</ins> (there are some goodies described in the end that will probably save you time - but you need the big picture first).


# Buckets

Buckets always live in a _COS Resource Instance_ - so you need to have one of those ready (you can create an instance [via crossplane](../resourcecontrollerv2/README.md), or the IBM Cloud UI or the IBM Cloud API).

Interestingly, buckets must have <ins>unique names</ins> across the IBM Cloud (this has implications in creating/importing buckets - arguably makes life easier).

# Creation of a bucket

The information you need to create a bucket is

* the __name__ you want to call it (note the uniqueness requirement above)
* either
  * the __CRN__ of the _IBM COS resource instance_  it will live in, __OR__
  * the __UID__ of the _IBM COS resource instance_ it will live in, 
  * the __name-in-kubernetes__ of the (kubernetes) resource instance it will live in
* the __geo/location__ of the bucket
  * note that only <ins>us-cold</ins> and <ins>us-standard</ins> are currently supported by the IBM Cloud API
* Optionally (if you want to encrypt it), the __CRN__ of the _(encryption) Root key_ you want to use
  * in order to do this you need to have associated the COS service <ins>associated with the Key Protect service</ins> (the one managing the key you want to use), beforehand. This is described [here](https://cloud.ibm.com/docs/key-protect?topic=key-protect-integrate-cos)
  
Once you have all the above - we will tell you later how to "read" this stuff programmatically - and provided your kubernetes has an authentication token to the IBM cloud (stored as a secret), you can create a bucket, by setting the following variables in your shell 

```shell
BUCKET_NAME=<there are format constraints - consult the documentatin of what is acceptable>
LOCATION=us-cold OR us-stanadard
DELETION_POLICY=Orphan OR Delete
RESOURCE_CONTROLLER_SERVICE_INSTANCE_CRN=... (you need either this or the next one; see below on how to get it)
KUBE_NAME_OF_ENCLOSING_INSTANCE=... (you need either this or the previous one; see below on how to get it)
ROOT_KEY_CRN=...optional (see info below on how to retrieve it)
ENCR_ALGORITHM=AES256 OR empty (if no encryption required)
```

...and running the following script

```shell
cat <<EOF | kubectl apply -f -
apiVersion: cos.ibmcloud.crossplane.io/v1alpha1
kind: Bucket
metadata:
  name: $BUCKET_NAME
  annotations:
    
spec:
  deletionPolicy: $DELETION_POLICY
  forProvider:
    bucket: $BUCKET_NAME
    ibmServiceInstanceID: '$RESOURCE_CONTROLLER_SERVICE_INSTANCE_CRN'
    ibmServiceInstanceIDRef: 
        name: '$KUBE_NAME_OF_ENCLOSING_INSTANCE'
    ibmServiceInstanceIDSelector:
    ibmSSEKpCustomerRootKeyCrn: '$ROOT_KEY_CRN'
    ibmSSEKpEncryptionAlgorithm: $ENCR_ALGORITHM
    locationConstraint: $LOCATION
  providerConfigRef:
    name: ibm-cloud
EOF
```

...or via filling in the info in one of the templates provided in this dir...

### How to get the info...

* For the case of an unencrypted bucket, all you need is <ins>one of</ins>
  * the __CRN__ of the hosting _COS Resource Instance_
    * you can get that via the steps described [here](../resourcecontrollerv2/README.md), _
  * the __UID__ of the hosting _COS Resource Instance_
    * you can get that via the steps described [here](../resourcecontrollerv2/README.md), 
  * the __name__ of the hosting _kube resource instance_
    * you can get this by via ```kubectl get resourceinstances```

* If you also want to encrypt the bucket, you need, in addition (as described - and linked - earlier)
  * a __Key Protect__ service
    * containing a __Root key__
    * ..and "associated" with the _COS Service instance_ 
  *  the <ins>CRN of the root key</ins>. You can get this via using the cloud terminal and running the following shell script (of which you need to configure only the first 2 lines)

```shell
KP_INSTANCE_NAME="<name of the Key Protect instance>"
ENCRYPTION_ROOT_KEY_NAME="<the name of the one you want to use>"

kp_service_instance_info=$(ibmcloud resource service-instance "$KP_INSTANCE_NAME" --id | grep crn)
kp_service_instance_crn=$(echo "$kp_service_instance_info" | cut -d' ' -f 1)
kp_service_instance_id=$(echo "$kp_service_instance_info" | cut -d' ' -f 2)
root_key_id=$(ibmcloud kp keys -i "$kp_service_instance_id" | grep "$ENCRYPTION_ROOT_KEY_NAME" | cut -d' ' -f 1)
ROOT_KEY_CRN=$(echo "$kp_service_instance_crn:key:$root_key_id" | sed 's/:::/:/')
echo "$ROOT_KEY_CRN"
```

(Note that you can run the above even from your own laptop, if you have 
* installed the `ibmcloud` utility,
* (if you are dealing with encryption...) installed the [key protect CLI](https://cloud.ibm.com/docs/key-protect?topic=key-protect-set-up-cli), 
* authenticated to the ibm cloud - via IBM login))


# Importing a bucket

Importing an IBM Cloud-hosted bucket in a kubernetes control plane does NOT go through our crossplane controller (although, once imported, crossplane's "default" idea is to attempt to "reconcile" it with the version in the IBM Cloud. But there is a caveat which will affect this default. More below).

The only difference from creating one is that you need to have an external name (ie the one in the IBM cloud) to a  YAML file like the one listed above. So this would make the script used above for creation look like

```shell
cat <<EOF | kubectl apply -f -
apiVersion: cos.ibmcloud.crossplane.io/v1alpha1
kind: Bucket
metadata:
  name: $BUCKET_NAME
  annotations:
        crossplane.io/external-name: "$BUCKET_NAME"       <-- only difference wrt creation is the presence of this line
spec:
  deletionPolicy: $DELETION_POLICY
  forProvider:
    bucket: $BUCKET_NAME
    ibmServiceInstanceID: '$RESOURCE_CONTROLLER_SERVICE_INSTANCE_CRN'
    ibmServiceInstanceIDRef: 
        name: '$KUBE_NAME_OF_ENCLOSING_INSTANCE'
    ibmServiceInstanceIDSelector:
    ibmSSEKpCustomerRootKeyCrn: '$ROOT_KEY_CRN'
    ibmSSEKpEncryptionAlgorithm: $ENCR_ALGORITHM
    locationConstraint: $LOCATION
  providerConfigRef:
    name: ibm-cloud
EOF
```

...the difference from the "creation" yaml being <ins>the non-empty "body" of the _annotations_ section</ins>

# Caveat

This is important as it will likely bite you (and soon, too). See [here](../../pkg/controller/cos/README.md)

# Extras

In this dir you will find
* several yaml file templates (some in subdirs) - the contents should be obvious. Note that there are examples of using both the _IBM cloud CRN_ and the _kube names_ of the "containing" resource instances
* a script that you can run to create/import a bucket. 
  * Useful, as it obviates the need to "manually" look up CRNs as described above - you only need to know the names of the things - it can figure out the rest.
  * ...can be run as follows (note that the 2 last parameters - in square brackets - affect encryption and hence are optional)
  
```./create-import-bucket.sh <bucket name> <bucket location> (--ri-cloud "<COS resource instance NAME in cloud>" | --ri-crossplane "<NAME of resource instance in crossplane>") (--create | --import) (--orphan | --delete) [<key-protect service name> <root key name>]```

Some options may not be obvious:

 `--create` vs `--import` is the difference between "create a bucket" and "import a bucket"
 
 `--delete` vs `--orphan` is the difference between a subsequent deletion "propagating to the IBM cloud" (hence the bucket gets deleted there, too) and "only happening in the kubernetes control plane" (hence the bucket in the IBM cloud gets "orphaned". If you feel sorry for it you can import it later).

eg

* ```./create-import-bucket.sh a-name us-cold --ri-cloud "IBM Cloud Name" --import --orphan```
* ```./create-import-bucket.sh a-name us-standard --ri-crossplane cos-name-in-crossplane --create --delete```

