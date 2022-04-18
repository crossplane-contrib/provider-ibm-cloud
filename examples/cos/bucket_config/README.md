# Bucket Configurations
Every bucket in the IBM cloud has a configuration (like files and their permissions). Hence

* You cannot create or delete a bucket configuration - you can only modify it (kind of like file permissions)
  * you can certainly delete it from the k8s control plane
    * hence the deletion policy <ins>should always be set to _Orphan_</ins>
* the name  of the bucket config is the same as that of the bucket
  * you can use either the _name_ or the _nameRef_ (the latter gets resolved to the former)

(more info about IBM cloud buckets [in crossplane](../bucket/README.md), or the IBM Cloud UI site or the IBM Cloud API docs).

### Discussion of the config spec 

(no need to paste into a file - it is in one of the example files provided)

The full spec is this 

```
apiVersion: cos.ibmcloud.crossplane.io/v1alpha1
kind: BucketConfig
metadata:
  name: configurable-1
spec:
  deletionPolicy: Orphan
  forProvider:
    name:
    nameRef: 
        name: configurable-1
    nameSelector:
    hardQuota: 201
    metricsMonitoring:
      usageMetricsEnabled: false
      requestMetricsEnabled: false
      metricsMonitoringCRN: "crn:v1:bluemix:public:sysdig-monitor:us-south:a/0b5a00334eaf9eb9339d2ab48f20d7f5:636a39a0-aeb2-424c-949e-546df6d79d42::"        
    activityTracking:
      readDataEvents: false
      writeDataEvents: false
      activityTrackerCRN: "crn:v1:bluemix:public:logdnaat:us-south:a/0b5a00334eaf9eb9339d2ab48f20d7f5:ad5c1154-4ebc-4c46-9a9f-030792b29138::"
    firewall:
        allowedIP: []
  providerConfigRef:
    name: ibm-cloud
```

Note that we specified the `nameRef` - not the `name`; this is the name of the bucket and it assumes there is a bucket with this name defined in the cluster. We could have used the `name` instead, and then no bucket-in-the-cluster would be necessary

Note also that there are not that many things that can be configured (the UI has more). This is the current state of the API we have to use; likely things will change.

...and a minimal (ie with only the compulsory fields) example:

```
apiVersion: cos.ibmcloud.crossplane.io/v1alpha1
kind: BucketConfig
metadata:
  name: configurable-1
spec:
  deletionPolicy: Orphan
  forProvider:
    name: configurable-1
  providerConfigRef:
    name: ibm-cloud
```

Things to keep in mind

* _deletionPolicy_ is ALWAYS _Orphan_
* after the first call to crossplane's `Observe`... whatever fields were left unpopulated on the k8s side will get their values from the IBM cloud (so the minimal example may "expand" during the next sync cycle)
* updates to the config in the IBM cloud happen via JSON patch. Hence, if you want to update the value of some fields only, create a yaml file with __values for only those fields__, and `kubeclt apply -f ...`
  * ...but if you leave the values empty, or do not enter them at all (in the code it is the same thing - those values are just 'nil') <ins>nothing will happen</ins>: the corresp fields in the IBM cloud <ins>WILL NOT</ins> get affected
  * CRNs are needed the first time around (essentially when creating the thing) - from then on you can leave them empty 
      * the documentation on the IBM cloud API says what a valid CRN looks like


### How to delete/disable things

* Setting the CRN of the _activity tracking_ or _metrics monitoring_ component to __"0"__ (note the quotes - without them the yaml validator thinks it is a number and chokes) will DISABLE the tracking/monitoring (as this is not a valid CRN)
   * ...it does not matter whether you have set the remaining fields to anything - the CRN wins
* Setting the `hardQuota` to __0__ (no quotes here - this is a legit number) will delete whatever quota the bucket has (essentially clearing the value)
* Setting the firewall's allowed IP list to empty ```allowedIP: []``` will disable it


### Examples
Please check the following 3 examples:

* A [partial update to the config](partial_update.yaml). Note that
  * leaving `ActivityConfig` empty will NOT disable it. <ins>You need to set the CRN to __"0"__</ins>, to disable it
* One that [deletes everything that can be deleted](delete_stuff.yaml)
* A [minimal example](minimal_aka_import.yaml) - essentially "importing" a bucket to k8s.