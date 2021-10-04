# Copyright 2021 The Crossplane Authors.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


bucket_name="$1"                     # To create or import

location="$2"                        # Either 'us-standard' or 'us-cold'

ibm_resource_id_type="$3"            # Either --ri-cloud or --ri-crossplane
ibm_resource_name_or_ref="$4"        # Name (in the IBM cloud OR in crossplane)
                                    
action="$5"                          # Either '--create' or '--import'

born_orphan="$6"                     # Either '--orphan' or '--delete'.
                                     # If '--orphan', then, on deletion the
                                     # bucket in the cloud does not get deleted

kp_instance_name="$7"                # Name of the key-protect instance
                                     # containing the - root - key used to
                                     # encrypt the bucket. Optional

encryption_root_key_name="$8"        # Name of the root key to protect the
                                     # bucket. To be used with the previous
                                     # param kp_instance_name, hence optional


# If we need encryption, we need to discover the CRN name of the root encryption
# key
if [[ "$kp_instance_name" != "" ]]; then
    kp_service_instance_info=$(ibmcloud resource service-instance "$kp_instance_name" --id | grep crn)
    rc=$?
    if [[ $rc -ne 0 ]]; then
        echo "Problem getting the info for resource '$kp_instance_name'. The 'ibmcloud resource service-instance...' command returned: $rc" >&2

        exit $rc
    fi
    
    kp_service_instance_crn=$(echo "$kp_service_instance_info" | cut -d' ' -f 1)
    kp_service_instance_id=$(echo "$kp_service_instance_info" | cut -d' ' -f 2)
    root_key_id=$(ibmcloud kp keys -i "$kp_service_instance_id" | grep "$encryption_root_key_name" | cut -d' ' -f 1)
    rc=$?
    if [[ $rc -ne 0 ]]; then
        echo "Problem retrieving the keys for service instance id '$kp_service_instance_id'. The 'ibmcloud kp keys...' command returned: $rc" >&2

        exit $rc
    fi
    ibmSSEKpCustomerRootKeyCrn=$(echo "$kp_service_instance_crn:key:$root_key_id" | sed 's/:::/:/')

    if [[ "$ibmSSEKpCustomerRootKeyCrn" == "" ]]; then
        echo "Could not retrieve the CRN of the root key" >&2
        
        exit 1
    fi
fi

# Settle on the deletion policy
if [[ "$born_orphan" == "--orphan" ]]; then
    deletion_policy=Orphan
elif [[ "$born_orphan" == "--delete" ]]; then
    deletion_policy=Delete
else
    echo "Could not understand the deletion policy: '$born_orphan'. Allowed options are: '--orphan' and '--delete'." >&2

    exit 2
fi

# Do we need encryption?
ibmSSEKpEncryptionAlgorithm=
if [[ "$ibmSSEKpCustomerRootKeyCrn" != "" ]]; then
    ibmSSEKpEncryptionAlgorithm=AES256
fi
    
# Annotations are used for importing only...
annotation_str=
if [[ "$action" == "--import" ]]; then
    annotation_str="crossplane.io/external-name: \"$bucket_name\"" 
elif [[ "$action" != "--create" ]]; then
    echo "Could not understand the 'action' policy: '$action'. Allowed options are '--create' and '--import'." >&2

    exit 2
fi
    
ibm_resource_instance_id=
ibm_resource_instance_ref=
if [[ "$ibm_resource_id_type" == "--ri-cloud" ]]; then
    # ...we need the "internal" id of the "resource instance" where the bucket will
    # live (or lives, if importing)
    info=$(ibmcloud resource service-instance "${ibm_resource_name_or_ref}")
    rc=$?
    if [[ $rc -ne 0 ]]; then
        echo "Problem getting the resource instance. The 'ibmcloud resource service-instance...' command returned: $rc" >&2

        exit $rc
    fi

    ibm_resource_instance_id=$(echo "$info" | grep ^ID | awk '{print $2}')
elif [[ "$ibm_resource_id_type" == "--ri-crossplane" ]]; then
    ibm_resource_instance_ref="$ibm_resource_name_or_ref"
else 
    echo "Could not understand the type of the resource-instance-id: '$ibm_resource_id_type'. Allowed options are '--ri-cloud' and '--ri-crossplane'." >&2

    exit 3
fi

# Got all we need - vamos!
tmp_file=$(uuidgen)

echo "
apiVersion: cos.ibmcloud.crossplane.io/v1alpha1
kind: Bucket
metadata:
  name: $bucket_name
  annotations:
    $annotation_str
spec:
  deletionPolicy: $deletion_policy
  forProvider:
    bucket: $bucket_name
    ibmServiceInstanceID: '$ibm_resource_instance_id'
    ibmServiceInstanceIDRef: 
        name: '$ibm_resource_instance_ref'
    ibmServiceInstanceIDSelector:
    ibmSSEKpCustomerRootKeyCrn: '$ibmSSEKpCustomerRootKeyCrn'
    ibmSSEKpEncryptionAlgorithm: '$ibmSSEKpEncryptionAlgorithm'
    locationConstraint: '$location'
  providerConfigRef:
    name: ibm-cloud
" > "$tmp_file"

kubectl apply -f "$tmp_file"
rc=$?
if [[ $rc -ne 0 ]]; then
    echo "Problem with 'kubectl apply -f \"$tmp_file\"; it returned: $rc" >&2
    echo
    echo "File contents" >&2
    echo "-------------" >&2
    
    cat $tmp_file >&2
else 
    cat $tmp_file
fi

rm -f "$tmp_file"

exit $rc