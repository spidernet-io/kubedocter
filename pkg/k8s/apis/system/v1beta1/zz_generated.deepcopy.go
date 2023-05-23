//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	spiderdoctor_spidernet_iov1beta1 "github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/spiderdoctor.spidernet.io/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DNSMetrics) DeepCopyInto(out *DNSMetrics) {
	*out = *in
	in.StartTime.DeepCopyInto(&out.StartTime)
	in.EndTime.DeepCopyInto(&out.EndTime)
	if in.Errors != nil {
		in, out := &in.Errors, &out.Errors
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Latencies = in.Latencies
	if in.ReplyCode != nil {
		in, out := &in.ReplyCode, &out.ReplyCode
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DNSMetrics.
func (in *DNSMetrics) DeepCopy() *DNSMetrics {
	if in == nil {
		return nil
	}
	out := new(DNSMetrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HttpAppHealthyTask) DeepCopyInto(out *HttpAppHealthyTask) {
	*out = *in
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	if in.Detail != nil {
		in, out := &in.Detail, &out.Detail
		*out = make([]HttpAppHealthyTaskDetail, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HttpAppHealthyTask.
func (in *HttpAppHealthyTask) DeepCopy() *HttpAppHealthyTask {
	if in == nil {
		return nil
	}
	out := new(HttpAppHealthyTask)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HttpAppHealthyTaskDetail) DeepCopyInto(out *HttpAppHealthyTaskDetail) {
	*out = *in
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	in.Metrics.DeepCopyInto(&out.Metrics)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HttpAppHealthyTaskDetail.
func (in *HttpAppHealthyTaskDetail) DeepCopy() *HttpAppHealthyTaskDetail {
	if in == nil {
		return nil
	}
	out := new(HttpAppHealthyTaskDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HttpMetrics) DeepCopyInto(out *HttpMetrics) {
	*out = *in
	in.StartTime.DeepCopyInto(&out.StartTime)
	in.EndTime.DeepCopyInto(&out.EndTime)
	if in.Errors != nil {
		in, out := &in.Errors, &out.Errors
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Latencies = in.Latencies
	if in.StatusCodes != nil {
		in, out := &in.StatusCodes, &out.StatusCodes
		*out = make(map[int]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HttpMetrics.
func (in *HttpMetrics) DeepCopy() *HttpMetrics {
	if in == nil {
		return nil
	}
	out := new(HttpMetrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LatencyDistribution) DeepCopyInto(out *LatencyDistribution) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LatencyDistribution.
func (in *LatencyDistribution) DeepCopy() *LatencyDistribution {
	if in == nil {
		return nil
	}
	out := new(LatencyDistribution)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetDNSTask) DeepCopyInto(out *NetDNSTask) {
	*out = *in
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	if in.Detail != nil {
		in, out := &in.Detail, &out.Detail
		*out = make([]NetDNSTaskDetail, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetDNSTask.
func (in *NetDNSTask) DeepCopy() *NetDNSTask {
	if in == nil {
		return nil
	}
	out := new(NetDNSTask)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetDNSTaskDetail) DeepCopyInto(out *NetDNSTaskDetail) {
	*out = *in
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	in.Metrics.DeepCopyInto(&out.Metrics)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetDNSTaskDetail.
func (in *NetDNSTaskDetail) DeepCopy() *NetDNSTaskDetail {
	if in == nil {
		return nil
	}
	out := new(NetDNSTaskDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetReachHealthyTask) DeepCopyInto(out *NetReachHealthyTask) {
	*out = *in
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	if in.Detail != nil {
		in, out := &in.Detail, &out.Detail
		*out = make([]NetReachHealthyTaskDetail, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetReachHealthyTask.
func (in *NetReachHealthyTask) DeepCopy() *NetReachHealthyTask {
	if in == nil {
		return nil
	}
	out := new(NetReachHealthyTask)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetReachHealthyTaskDetail) DeepCopyInto(out *NetReachHealthyTaskDetail) {
	*out = *in
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	in.Metrics.DeepCopyInto(&out.Metrics)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetReachHealthyTaskDetail.
func (in *NetReachHealthyTaskDetail) DeepCopy() *NetReachHealthyTaskDetail {
	if in == nil {
		return nil
	}
	out := new(NetReachHealthyTaskDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PluginReport) DeepCopyInto(out *PluginReport) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PluginReport.
func (in *PluginReport) DeepCopy() *PluginReport {
	if in == nil {
		return nil
	}
	out := new(PluginReport)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PluginReport) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PluginReportList) DeepCopyInto(out *PluginReportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PluginReport, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PluginReportList.
func (in *PluginReportList) DeepCopy() *PluginReportList {
	if in == nil {
		return nil
	}
	out := new(PluginReportList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PluginReportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PluginReportSpec) DeepCopyInto(out *PluginReportSpec) {
	*out = *in
	if in.FailedRoundNumber != nil {
		in, out := &in.FailedRoundNumber, &out.FailedRoundNumber
		*out = make([]int64, len(*in))
		copy(*out, *in)
	}
	if in.Report != nil {
		in, out := &in.Report, &out.Report
		*out = new([]Report)
		if **in != nil {
			in, out := *in, *out
			*out = make([]Report, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PluginReportSpec.
func (in *PluginReportSpec) DeepCopy() *PluginReportSpec {
	if in == nil {
		return nil
	}
	out := new(PluginReportSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Report) DeepCopyInto(out *Report) {
	*out = *in
	if in.FailedReason != nil {
		in, out := &in.FailedReason, &out.FailedReason
		*out = new(string)
		**out = **in
	}
	in.StartTimeStamp.DeepCopyInto(&out.StartTimeStamp)
	in.EndTimeStamp.DeepCopyInto(&out.EndTimeStamp)
	if in.NetReachHealthyTaskSpec != nil {
		in, out := &in.NetReachHealthyTaskSpec, &out.NetReachHealthyTaskSpec
		*out = new(spiderdoctor_spidernet_iov1beta1.NetReachHealthySpec)
		(*in).DeepCopyInto(*out)
	}
	if in.NetReachHealthyTask != nil {
		in, out := &in.NetReachHealthyTask, &out.NetReachHealthyTask
		*out = new(NetReachHealthyTask)
		(*in).DeepCopyInto(*out)
	}
	if in.HttpAppHealthyTaskSpec != nil {
		in, out := &in.HttpAppHealthyTaskSpec, &out.HttpAppHealthyTaskSpec
		*out = new(spiderdoctor_spidernet_iov1beta1.HttpAppHealthySpec)
		(*in).DeepCopyInto(*out)
	}
	if in.HttpAppHealthyTask != nil {
		in, out := &in.HttpAppHealthyTask, &out.HttpAppHealthyTask
		*out = new(HttpAppHealthyTask)
		(*in).DeepCopyInto(*out)
	}
	if in.NetDNSTaskSpec != nil {
		in, out := &in.NetDNSTaskSpec, &out.NetDNSTaskSpec
		*out = new(spiderdoctor_spidernet_iov1beta1.NetdnsSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.NetDNSTask != nil {
		in, out := &in.NetDNSTask, &out.NetDNSTask
		*out = new(NetDNSTask)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Report.
func (in *Report) DeepCopy() *Report {
	if in == nil {
		return nil
	}
	out := new(Report)
	in.DeepCopyInto(out)
	return out
}
