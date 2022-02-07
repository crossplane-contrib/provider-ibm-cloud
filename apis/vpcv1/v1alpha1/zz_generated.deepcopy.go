// +build !ignore_autogenerated

/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IP) DeepCopyInto(out *IP) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IP.
func (in *IP) DeepCopy() *IP {
	if in == nil {
		return nil
	}
	out := new(IP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkACLReference) DeepCopyInto(out *NetworkACLReference) {
	*out = *in
	if in.Deleted != nil {
		in, out := &in.Deleted, &out.Deleted
		*out = new(NetworkACLReferenceDeleted)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkACLReference.
func (in *NetworkACLReference) DeepCopy() *NetworkACLReference {
	if in == nil {
		return nil
	}
	out := new(NetworkACLReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkACLReferenceDeleted) DeepCopyInto(out *NetworkACLReferenceDeleted) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkACLReferenceDeleted.
func (in *NetworkACLReferenceDeleted) DeepCopy() *NetworkACLReferenceDeleted {
	if in == nil {
		return nil
	}
	out := new(NetworkACLReferenceDeleted)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupIdentityAlsoByID) DeepCopyInto(out *ResourceGroupIdentityAlsoByID) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupIdentityAlsoByID.
func (in *ResourceGroupIdentityAlsoByID) DeepCopy() *ResourceGroupIdentityAlsoByID {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupIdentityAlsoByID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroupReference) DeepCopyInto(out *ResourceGroupReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroupReference.
func (in *ResourceGroupReference) DeepCopy() *ResourceGroupReference {
	if in == nil {
		return nil
	}
	out := new(ResourceGroupReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingTableReference) DeepCopyInto(out *RoutingTableReference) {
	*out = *in
	out.Deleted = in.Deleted
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingTableReference.
func (in *RoutingTableReference) DeepCopy() *RoutingTableReference {
	if in == nil {
		return nil
	}
	out := new(RoutingTableReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingTableReferenceDeleted) DeepCopyInto(out *RoutingTableReferenceDeleted) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingTableReferenceDeleted.
func (in *RoutingTableReferenceDeleted) DeepCopy() *RoutingTableReferenceDeleted {
	if in == nil {
		return nil
	}
	out := new(RoutingTableReferenceDeleted)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupReference) DeepCopyInto(out *SecurityGroupReference) {
	*out = *in
	out.Deleted = in.Deleted
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupReference.
func (in *SecurityGroupReference) DeepCopy() *SecurityGroupReference {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityGroupReferenceDeleted) DeepCopyInto(out *SecurityGroupReferenceDeleted) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityGroupReferenceDeleted.
func (in *SecurityGroupReferenceDeleted) DeepCopy() *SecurityGroupReferenceDeleted {
	if in == nil {
		return nil
	}
	out := new(SecurityGroupReferenceDeleted)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPC) DeepCopyInto(out *VPC) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPC.
func (in *VPC) DeepCopy() *VPC {
	if in == nil {
		return nil
	}
	out := new(VPC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VPC) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCList) DeepCopyInto(out *VPCList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VPC, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCList.
func (in *VPCList) DeepCopy() *VPCList {
	if in == nil {
		return nil
	}
	out := new(VPCList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VPCList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCObservation) DeepCopyInto(out *VPCObservation) {
	*out = *in
	if in.CreatedAt != nil {
		in, out := &in.CreatedAt, &out.CreatedAt
		*out = (*in).DeepCopy()
	}
	if in.CseSourceIps != nil {
		in, out := &in.CseSourceIps, &out.CseSourceIps
		*out = make([]VpccseSourceIP, len(*in))
		copy(*out, *in)
	}
	in.DefaultNetworkACL.DeepCopyInto(&out.DefaultNetworkACL)
	in.DefaultRoutingTable.DeepCopyInto(&out.DefaultRoutingTable)
	out.DefaultSecurityGroup = in.DefaultSecurityGroup
	out.ResourceGroup = in.ResourceGroup
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCObservation.
func (in *VPCObservation) DeepCopy() *VPCObservation {
	if in == nil {
		return nil
	}
	out := new(VPCObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCParameters) DeepCopyInto(out *VPCParameters) {
	*out = *in
	if in.AddressPrefixManagement != nil {
		in, out := &in.AddressPrefixManagement, &out.AddressPrefixManagement
		*out = new(string)
		**out = **in
	}
	if in.ClassicAccess != nil {
		in, out := &in.ClassicAccess, &out.ClassicAccess
		*out = new(bool)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.ResourceGroup != nil {
		in, out := &in.ResourceGroup, &out.ResourceGroup
		*out = new(ResourceGroupIdentityAlsoByID)
		**out = **in
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCParameters.
func (in *VPCParameters) DeepCopy() *VPCParameters {
	if in == nil {
		return nil
	}
	out := new(VPCParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCSpec) DeepCopyInto(out *VPCSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCSpec.
func (in *VPCSpec) DeepCopy() *VPCSpec {
	if in == nil {
		return nil
	}
	out := new(VPCSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCStatus) DeepCopyInto(out *VPCStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCStatus.
func (in *VPCStatus) DeepCopy() *VPCStatus {
	if in == nil {
		return nil
	}
	out := new(VPCStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VpccseSourceIP) DeepCopyInto(out *VpccseSourceIP) {
	*out = *in
	out.IP = in.IP
	out.Zone = in.Zone
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VpccseSourceIP.
func (in *VpccseSourceIP) DeepCopy() *VpccseSourceIP {
	if in == nil {
		return nil
	}
	out := new(VpccseSourceIP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneReference) DeepCopyInto(out *ZoneReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneReference.
func (in *ZoneReference) DeepCopy() *ZoneReference {
	if in == nil {
		return nil
	}
	out := new(ZoneReference)
	in.DeepCopyInto(out)
	return out
}
