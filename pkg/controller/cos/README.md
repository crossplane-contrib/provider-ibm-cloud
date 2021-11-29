# Buckets

Buckets, once created, cannot be modified (eg if encrypted, they cannot be unencrypted. Or moved to a different geography). So the crossplane "Update" call does not do anything really... (because the spec is "write once" and the "Observe" reports always that things are up-to-date).

The problem this causes is that if you "import" a bucket - but while doing so you give the "wrong" (ie not-reflecting-the-state-of-affairs-in-the-IBM-cloud) spec, in the YAML file (eg the bucket is encrypted but you import is as unencrypted, or with a different location), ... the "copy" you-end-up-with-in-crossplane will have a different _ForProvider_ spec than the one in the IBM cloud. Keep in mind that

* our crossplane controller does NOT get involved in the "import"
    -  of course, subsequently, the _Observe(...)_ method will be called
* ...if the _Observe(..)_ were to report "not in sync", then the _Update(...)_ would be invoked
    - but it does not do anything currently, nor could it - even if it wanted
    - ...so to prevent continuous _Update(..)_ invocations, we just have _Observe(...)_ report "all good"
  
We can probably have the _Observe(...)_ method report this inconsistency in another way (eg via "status" - certainly not via telling crossplane "things are not in sync -> please call _Update(..)_ ASAP"), but we decided to not do it, for now.

More about buckets [here](../../../examples/cos/bucket/README.md)

# Bucket configurations

Read all about them [here](../../../examples/cos/bucket_config/README.md)