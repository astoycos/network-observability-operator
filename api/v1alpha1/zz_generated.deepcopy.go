// +build !ignore_autogenerated

/*
Copyright 2021.

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
func (in *FlowCollector) DeepCopyInto(out *FlowCollector) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollector.
func (in *FlowCollector) DeepCopy() *FlowCollector {
	if in == nil {
		return nil
	}
	out := new(FlowCollector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlowCollector) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorGoflowKube) DeepCopyInto(out *FlowCollectorGoflowKube) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorGoflowKube.
func (in *FlowCollectorGoflowKube) DeepCopy() *FlowCollectorGoflowKube {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorGoflowKube)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorIPFIX) DeepCopyInto(out *FlowCollectorIPFIX) {
	*out = *in
	out.CacheActiveTimeout = in.CacheActiveTimeout
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorIPFIX.
func (in *FlowCollectorIPFIX) DeepCopy() *FlowCollectorIPFIX {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorIPFIX)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorList) DeepCopyInto(out *FlowCollectorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FlowCollector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorList.
func (in *FlowCollectorList) DeepCopy() *FlowCollectorList {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlowCollectorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorLoki) DeepCopyInto(out *FlowCollectorLoki) {
	*out = *in
	out.BatchWait = in.BatchWait
	out.MinBackoff = in.MinBackoff
	out.MaxBackoff = in.MaxBackoff
	if in.StaticLabels != nil {
		in, out := &in.StaticLabels, &out.StaticLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorLoki.
func (in *FlowCollectorLoki) DeepCopy() *FlowCollectorLoki {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorLoki)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorSpec) DeepCopyInto(out *FlowCollectorSpec) {
	*out = *in
	out.IPFIX = in.IPFIX
	out.GoflowKube = in.GoflowKube
	in.Loki.DeepCopyInto(&out.Loki)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorSpec.
func (in *FlowCollectorSpec) DeepCopy() *FlowCollectorSpec {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowCollectorStatus) DeepCopyInto(out *FlowCollectorStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowCollectorStatus.
func (in *FlowCollectorStatus) DeepCopy() *FlowCollectorStatus {
	if in == nil {
		return nil
	}
	out := new(FlowCollectorStatus)
	in.DeepCopyInto(out)
	return out
}