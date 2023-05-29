// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package netdns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"

	k8sObjManager "github.com/spidernet-io/spiderdoctor/pkg/k8ObjManager"
	crd "github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/spiderdoctor.spidernet.io/v1beta1"
	"github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/system/v1beta1"
	"github.com/spidernet-io/spiderdoctor/pkg/loadRequest/loadDns"
	"github.com/spidernet-io/spiderdoctor/pkg/lock"
	"github.com/spidernet-io/spiderdoctor/pkg/pluginManager/types"
)

func ParseSuccessCondition(successCondition *crd.NetSuccessCondition, metricResult *v1beta1.DNSMetrics) (failureReason string, err error) {
	switch {
	case successCondition.SuccessRate != nil && float64(metricResult.SuccessCounts)/float64(metricResult.SuccessCounts) < *(successCondition.SuccessRate):
		failureReason = fmt.Sprintf("Success Rate %v is lower than request %v", float64(metricResult.SuccessCounts)/float64(metricResult.SuccessCounts), *(successCondition.SuccessRate))
	case successCondition.MeanAccessDelayInMs != nil && int64(metricResult.Latencies.Mean) > *(successCondition.MeanAccessDelayInMs):
		failureReason = fmt.Sprintf("mean delay %v ms is bigger than request %v ms", metricResult.Latencies.Mean, *(successCondition.MeanAccessDelayInMs))
	default:
		failureReason = ""
		err = nil
	}

	return
}

func SendRequestAndReport(logger *zap.Logger, targetName string, req *loadDns.DnsRequestData, successCondition *crd.NetSuccessCondition) (failureReason string, report v1beta1.NetDNSTaskDetail) {
	report.TargetName = targetName
	report.TargetServer = req.DnsServerAddr
	report.TargetProtocol = string(req.Protocol)

	result, err := loadDns.DnsRequest(logger, req)
	if err != nil {
		logger.Sugar().Errorf("internal error for target %v, error=%v", req.DnsServerAddr, err)
		report.FailureReason = pointer.String(err.Error())
		return
	}

	report.MeanDelay = result.Latencies.Mean
	report.SucceedRate = float64(result.SuccessCounts) / float64(result.RequestCounts)

	failureReason, err = ParseSuccessCondition(successCondition, result)
	if err != nil {
		logger.Sugar().Errorf("internal error for target %v, error=%v", req.DnsServerAddr, err)
		report.FailureReason = pointer.String(err.Error())
		return
	}

	// generate report
	// notice , upper case for first character of key, or else fail to parse json
	report.Metrics = *result
	report.FailureReason = pointer.String(failureReason)
	if report.FailureReason == nil {
		report.Succeed = true
		logger.Sugar().Infof("succeed to test %v", req.DnsServerAddr)
	} else {
		report.Succeed = false
		logger.Sugar().Warnf("failed to test %v", req.DnsServerAddr)
	}

	return
}

type testTarget struct {
	Name    string
	Request *loadDns.DnsRequestData
}

func (s *PluginNetDns) AgentExecuteTask(logger *zap.Logger, ctx context.Context, obj runtime.Object) (finalfailureReason string, finalReport types.Task, err error) {
	finalfailureReason = ""

	instance, ok := obj.(*crd.Netdns)
	if !ok {
		msg := "failed to get instance"
		logger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	logger.Sugar().Infof("plugin implement task round, instance=%+v", instance)

	var testTargetList []*testTarget
	var server string

	// Choose whether to request typeA or typeAAAA based on the address type of the server
	if instance.Spec.Target.NetDnsTargetUser != nil {
		server = net.JoinHostPort(*instance.Spec.Target.NetDnsTargetUser.Server, strconv.Itoa(*instance.Spec.Target.NetDnsTargetUser.Port))
		ip := net.ParseIP(*instance.Spec.Target.NetDnsTargetUser.Server)
		if ip.To4() != nil {
			testTargetList = append(testTargetList, &testTarget{Name: "typeA_" + server + "_" + instance.Spec.Request.Domain, Request: &loadDns.DnsRequestData{
				Protocol:              loadDns.RequestProtocol(*instance.Spec.Target.Protocol),
				DnsType:               dns.TypeA,
				TargetDomain:          instance.Spec.Request.Domain,
				DnsServerAddr:         server,
				PerRequestTimeoutInMs: int(*instance.Spec.Request.PerRequestTimeoutInMS),
				Qps:                   int(*instance.Spec.Request.QPS),
				DurationInSecond:      int(*instance.Spec.Request.DurationInSecond),
			}})
		} else {
			testTargetList = append(testTargetList, &testTarget{Name: "typeAAAA_" + server + "_" + instance.Spec.Request.Domain, Request: &loadDns.DnsRequestData{
				Protocol:              loadDns.RequestProtocol(*instance.Spec.Target.Protocol),
				DnsType:               dns.TypeAAAA,
				TargetDomain:          instance.Spec.Request.Domain,
				DnsServerAddr:         server,
				PerRequestTimeoutInMs: int(*instance.Spec.Request.PerRequestTimeoutInMS),
				Qps:                   int(*instance.Spec.Request.QPS),
				DurationInSecond:      int(*instance.Spec.Request.DurationInSecond),
			}})
		}
	}

	if instance.Spec.Target.NetDnsTargetDns != nil {
		// When DNS service is not specified, search for DNS services within the cluster
		if instance.Spec.Target.NetDnsTargetDns.ServiceNamespacedName == nil {
			dnsServiceIPs, err := k8sObjManager.GetK8sObjManager().ListServicesDnsIP(ctx)
			if err != nil {
				finalfailureReason = fmt.Sprintf("ListServicesDnsIP err: %v", err)
			}
			logger.Sugar().Infof("dnsServiceIPs %s", dnsServiceIPs)
			for _, serviceIP := range dnsServiceIPs {
				ip := net.ParseIP(serviceIP)
				server = net.JoinHostPort(serviceIP, "53")
				if ip.To4() != nil && *instance.Spec.Target.NetDnsTargetDns.TestIPv4 {
					testTargetList = append(testTargetList, &testTarget{Name: "typeA_" + server + "_" + instance.Spec.Request.Domain, Request: &loadDns.DnsRequestData{
						Protocol:              loadDns.RequestProtocol(*instance.Spec.Target.Protocol),
						DnsType:               dns.TypeA,
						TargetDomain:          instance.Spec.Request.Domain,
						DnsServerAddr:         server,
						PerRequestTimeoutInMs: int(*instance.Spec.Request.PerRequestTimeoutInMS),
						Qps:                   int(*instance.Spec.Request.QPS),
						DurationInSecond:      int(*instance.Spec.Request.DurationInSecond),
					}})
				} else if ip.To4() == nil && *instance.Spec.Target.NetDnsTargetDns.TestIPv6 {
					testTargetList = append(testTargetList, &testTarget{Name: "typeAAAA_" + server + "_" + instance.Spec.Request.Domain, Request: &loadDns.DnsRequestData{
						Protocol:              loadDns.RequestProtocol(*instance.Spec.Target.Protocol),
						DnsType:               dns.TypeAAAA,
						TargetDomain:          instance.Spec.Request.Domain,
						DnsServerAddr:         server,
						PerRequestTimeoutInMs: int(*instance.Spec.Request.PerRequestTimeoutInMS),
						Qps:                   int(*instance.Spec.Request.QPS),
						DurationInSecond:      int(*instance.Spec.Request.DurationInSecond),
					}})
				}
			}
		} else {
			// eg: kube-system/coredns
			namespacedName := strings.Split(*instance.Spec.Target.NetDnsTargetDns.ServiceNamespacedName, "/")
			dnsServices, err := k8sObjManager.GetK8sObjManager().GetService(ctx, namespacedName[1], namespacedName[0])
			if err != nil {
				finalfailureReason = fmt.Sprintf("GetService name: %s namespace: %s err: %v", namespacedName[1], namespacedName[0], err)
			}
			for _, serviceIP := range dnsServices.Spec.ClusterIPs {
				ip := net.ParseIP(serviceIP)
				server = net.JoinHostPort(serviceIP, "53")
				if ip.To4() != nil && *instance.Spec.Target.NetDnsTargetDns.TestIPv4 {
					testTargetList = append(testTargetList, &testTarget{Name: "typeA_" + server + "_" + instance.Spec.Request.Domain, Request: &loadDns.DnsRequestData{
						Protocol:              loadDns.RequestProtocol(*instance.Spec.Target.Protocol),
						DnsType:               dns.TypeA,
						TargetDomain:          instance.Spec.Request.Domain,
						DnsServerAddr:         server,
						PerRequestTimeoutInMs: int(*instance.Spec.Request.PerRequestTimeoutInMS),
						Qps:                   int(*instance.Spec.Request.QPS),
						DurationInSecond:      int(*instance.Spec.Request.DurationInSecond),
					}})
				} else if ip.To4() == nil && *instance.Spec.Target.NetDnsTargetDns.TestIPv6 {
					testTargetList = append(testTargetList, &testTarget{Name: "typeAAAA_" + server + "_" + instance.Spec.Request.Domain, Request: &loadDns.DnsRequestData{
						Protocol:              loadDns.RequestProtocol(*instance.Spec.Target.Protocol),
						DnsType:               dns.TypeAAAA,
						TargetDomain:          instance.Spec.Request.Domain,
						DnsServerAddr:         server,
						PerRequestTimeoutInMs: int(*instance.Spec.Request.PerRequestTimeoutInMS),
						Qps:                   int(*instance.Spec.Request.QPS),
						DurationInSecond:      int(*instance.Spec.Request.DurationInSecond),
					}})
				}
			}
		}

	}

	var reportList []v1beta1.NetDNSTaskDetail

	var wg sync.WaitGroup
	var l lock.Mutex
	for _, item := range testTargetList {
		wg.Add(1)
		go func(wg *sync.WaitGroup, l *lock.Mutex, t testTarget) {
			logger.Sugar().Debugf("implement test %v, request %v ", t.Name, *t.Request)
			failureReason, itemReport := SendRequestAndReport(logger, t.Name, t.Request, instance.Spec.SuccessCondition)
			if failureReason != "" {
				finalfailureReason = fmt.Sprintf("test %v: %v", t.Name, failureReason)
			}
			l.Lock()
			reportList = append(reportList, itemReport)
			l.Unlock()
			wg.Done()
		}(&wg, &l, *item)
	}
	wg.Wait()

	logger.Sugar().Infof("plugin finished all http request tests")

	// ----------------------- aggregate report
	task := &v1beta1.NetDNSTask{}
	task.Detail = reportList
	task.TargetType = "spiderdoctor agent"
	task.TargetNumber = int64(len(testTargetList))
	if len(finalfailureReason) > 0 {
		logger.Sugar().Errorf("plugin finally failed, %v", finalfailureReason)
		task.FailureReason = pointer.String(finalfailureReason)
		task.Succeed = false
	} else {
		task.Succeed = true
	}

	return finalfailureReason, task, err
}

func (s *PluginNetDns) SetReportWithTask(report *v1beta1.Report, crdSpec interface{}, task types.Task) error {
	netdnsSpec, ok := crdSpec.(*crd.NetdnsSpec)
	if !ok {
		return fmt.Errorf("the given crd spec %#v doesn't match NetdnsSpec", crdSpec)
	}

	netDNSTask, ok := task.(*v1beta1.NetDNSTask)
	if !ok {
		return fmt.Errorf("task type %v doesn't match NetDNSTask", task.KindTask())
	}

	report.NetDNSTaskSpec = netdnsSpec
	report.NetDNSTask = netDNSTask
	return nil
}
