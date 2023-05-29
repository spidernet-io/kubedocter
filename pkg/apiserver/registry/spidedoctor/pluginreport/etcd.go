// Copyright 2023 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package pluginreport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	"k8s.io/klog/v2"

	"github.com/spidernet-io/spiderdoctor/pkg/apiserver/registry"
	crd "github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/spiderdoctor.spidernet.io/v1beta1"
	"github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/system/v1beta1"
	"github.com/spidernet-io/spiderdoctor/pkg/k8s/client/clientset/versioned"
)

const dir = "/report"

func NewREST(clientSet *versioned.Clientset, scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*registry.REST, error) {
	strategy := NewStrategy(scheme)

	restOptions, err := optsGetter.GetRESTOptions(v1beta1.Resource("pluginreports"))
	if nil != err {
		return nil, err
	}

	dryRunnableStorage, destroyFunc := NewStorage(clientSet, restOptions)
	store := &genericregistry.Store{
		NewFunc:     func() runtime.Object { return &v1beta1.PluginReport{} },
		NewListFunc: func() runtime.Object { return &v1beta1.PluginReportList{} },
		KeyRootFunc: func(ctx context.Context) string {
			return restOptions.ResourcePrefix
		},
		KeyFunc: func(ctx context.Context, name string) (string, error) {
			return genericregistry.NoNamespaceKeyFunc(ctx, restOptions.ResourcePrefix, name)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*v1beta1.PluginReport).Name, nil
		},
		DefaultQualifiedResource: v1beta1.Resource("pluginreports"),
		PredicateFunc:            MatchPluginReport,

		CreateStrategy:          strategy,
		UpdateStrategy:          strategy,
		DeleteStrategy:          strategy,
		EnableGarbageCollection: true,

		Storage:     dryRunnableStorage,
		DestroyFunc: destroyFunc,

		//TableConvertor: printers.NewTableGenerator(v1beta1.Resource("pluginreports")),
	}

	return &registry.REST{Store: store}, nil
}

func NewStorage(clientSet *versioned.Clientset, restOptions generic.RESTOptions) (genericregistry.DryRunnableStorage, factory.DestroyFunc) {
	dryRunnableStorage := genericregistry.DryRunnableStorage{
		Storage: &pluginReportStorage{
			clientSet: clientSet,
		},
		Codec: restOptions.StorageConfig.Codec,
	}

	return dryRunnableStorage, func() {}
}

var _ storage.Interface = &pluginReportStorage{}

type pluginReportStorage struct {
	clientSet    *versioned.Clientset
	resourceName string
}

func (p pluginReportStorage) Versioner() storage.Versioner {
	return storage.APIObjectVersioner{}
}

func (p pluginReportStorage) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	return fmt.Errorf("create API not implement")
}

func (p pluginReportStorage) Delete(ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions, validateDeletion storage.ValidateObjectFunc, cachedExistingObject runtime.Object) error {
	return fmt.Errorf("delete API not implement")
}

func (p pluginReportStorage) Watch(ctx context.Context, key string, opts storage.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("watch API not implement")

}

func (p pluginReportStorage) Get(ctx context.Context, key string, opts storage.GetOptions, objPtr runtime.Object) error {
	klog.Infof("Get called with key: %v on resource %v\n", key, p.resourceName)

	var taskStatus *crd.TaskStatus
	var taskType string

	_, name, err := NamespaceAndNameFromKey(key, false)
	if nil != err {
		return err
	}

	// TODO (Icarus9913): we need options to specify which CRD that we are looking for.
	netdns, err := p.clientSet.SpiderdoctorV1beta1().Netdnses().Get(ctx, name, metav1.GetOptions{})
	if nil != err {
		if errors.IsNotFound(err) {
			klog.Infof("no NetDNS %s found", name)
		} else {
			return fmt.Errorf("failed to get NetDNS %s, error: %w", name, err)
		}
	} else {
		fmt.Printf("succeed to get NetDNS %s\n", name)
		taskStatus = netdns.Status.DeepCopy()
		taskType = v1beta1.NetDNSTaskName
	}

	netReachHealthy, err := p.clientSet.SpiderdoctorV1beta1().NetReachHealthies().Get(ctx, name, metav1.GetOptions{})
	if nil != err {
		if errors.IsNotFound(err) {
			klog.Infof("no NetReachHealthy %s found", name)
		} else {
			return fmt.Errorf("failed to get NetReachHealthy %s, error: %w", name, err)
		}
	} else {
		fmt.Printf("succeed to get NetReachHealthy %s\n", name)
		taskStatus = netReachHealthy.Status.DeepCopy()
		taskType = v1beta1.NetReachHealthyTaskName
	}

	httpAppHealthy, err := p.clientSet.SpiderdoctorV1beta1().HttpAppHealthies().Get(ctx, name, metav1.GetOptions{})
	if nil != err {
		if errors.IsNotFound(err) {
			klog.Infof("no HttpAppHealthy %s found", name)
		} else {
			return fmt.Errorf("failed to get HttpAppHealthy %s, error: %w", name, err)
		}
	} else {
		fmt.Printf("succeed to get HttpAppHealthy %s\n", name)
		taskStatus = httpAppHealthy.Status.DeepCopy()
		taskType = v1beta1.HttpAppHealthyTaskName
	}

	if taskStatus == nil {
		return fmt.Errorf("no crd instance %s found", name)
	}
	var status string
	if taskStatus.Finish {
		status = "Finished"
	} else {
		status = "NotFinished"
	}

	var toTalRoundNumber, finishedRoundNumber int64
	if taskStatus.ExpectedRound != nil {
		toTalRoundNumber = *taskStatus.ExpectedRound
	}
	if taskStatus.DoneRound != nil {
		finishedRoundNumber = *taskStatus.DoneRound
	}

	readDir, err := os.ReadDir(dir)
	if nil != err {
		return fmt.Errorf("failed to read directory %s, error: %w", dir, err)
	}
	var fileNameList []string
	for _, item := range readDir {
		if item.IsDir() {
			continue
		}

		if strings.Contains(item.Name(), name) && !strings.Contains(item.Name(), summary) {
			fileNameList = append(fileNameList, item.Name())
		}
	}

	getReports, latestRoundNumber, err := p.getLatestRoundReports(name, fileNameList)
	if nil != err {
		return fmt.Errorf("failed to get latest round reports: %w", err)
	}
	if getReports == nil {
		return fmt.Errorf("no '%s' reports found", name)
	}

	pluginReport := objPtr.(*v1beta1.PluginReport)
	pluginReport.Spec = v1beta1.PluginReportSpec{
		TaskName:            name,
		TaskType:            taskType,
		ToTalRoundNumber:    toTalRoundNumber,
		FinishedRoundNumber: finishedRoundNumber,
		FailedRoundNumber:   nil,
		Status:              status,
		ReportRoundNumber:   latestRoundNumber,
		Report:              getReports,
	}

	pluginReport.Name = name
	pluginReport.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1beta1.GroupName,
		Version: v1beta1.V1betaVersion,
		Kind:    v1beta1.KindPluginReport,
	})

	return nil
}

const summary = "summary"

func (p pluginReportStorage) GetList(ctx context.Context, key string, opts storage.ListOptions, listObj runtime.Object) error {
	readDir, err := os.ReadDir(dir)
	if nil != err {
		return fmt.Errorf("failed to read directory %s, error: %w", dir, err)
	}
	var fileNameList []string
	for _, item := range readDir {
		if item.IsDir() {
			continue
		}
		fileNameList = append(fileNameList, item.Name())
	}

	pluginReportList := listObj.(*v1beta1.PluginReportList)
	var resList []runtime.Object

	{
		netDNSPluginReports, err := p.getNetDNSPluginReports(ctx, fileNameList)
		if nil != err {
			return err
		}
		for i := range netDNSPluginReports {
			resList = append(resList, netDNSPluginReports[i].DeepCopy())
		}
	}

	{
		httpAppHealthyReports, err := p.getHttpAppHealthyReports(ctx, fileNameList)
		if nil != err {
			return err
		}
		for i := range httpAppHealthyReports {
			resList = append(resList, httpAppHealthyReports[i].DeepCopy())
		}
	}

	{
		netReachHealthyReports, err := p.getNetReachHealthyReports(ctx, fileNameList)
		if nil != err {
			return err
		}
		for i := range netReachHealthyReports {
			resList = append(resList, netReachHealthyReports[i].DeepCopy())
		}
	}

	err = meta.SetList(pluginReportList, resList)
	if nil != err {
		return err
	}

	pluginReportList.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
		Group:   v1beta1.GroupName,
		Version: v1beta1.V1betaVersion,
		Kind:    v1beta1.KindPluginReportList,
	})

	return nil
}

func (p pluginReportStorage) GuaranteedUpdate(ctx context.Context, key string, destination runtime.Object, ignoreNotFound bool, preconditions *storage.Preconditions, tryUpdate storage.UpdateFunc, cachedExistingObject runtime.Object) error {
	return fmt.Errorf("GuaranteedUpdate API not implement")
}

func (p pluginReportStorage) Count(key string) (int64, error) {
	return 0, fmt.Errorf("Count not supported for key: %s", key)
}

func (p pluginReportStorage) getLatestRoundReports(key string, fileNameList []string) (*[]v1beta1.Report, int64, error) {
	var reports []v1beta1.Report
	for _, netDNSFileName := range fileNameList {
		split := strings.Split(netDNSFileName, "_")
		if len(split) != 5 {
			klog.Infof("unrecognized file %s", netDNSFileName)
			continue
		}

		if key == split[1] {
			file, err := os.Open(path.Join(dir, netDNSFileName))
			if nil != err {
				return nil, -1, err
			}
			readAll, err := io.ReadAll(file)
			if nil != err {
				return nil, -1, err
			}

			report := v1beta1.Report{}
			err = json.Unmarshal(readAll, &report)
			if nil != err {
				return nil, -1, err
			}
			reports = append(reports, report)
		}
	}

	var latestRoundNumber int64
	result := func() *[]v1beta1.Report {
		var latestReports []v1beta1.Report

		for _, tmpReport := range reports {
			if tmpReport.RoundNumber > latestRoundNumber {
				latestRoundNumber = tmpReport.RoundNumber
				latestReports = []v1beta1.Report{tmpReport}
			} else if tmpReport.RoundNumber == latestRoundNumber {
				latestReports = append(latestReports, tmpReport)
			} else {
				continue
			}
		}
		if len(latestReports) == 0 {
			return nil
		}

		return &latestReports
	}()

	return result, latestRoundNumber, nil
}

func (p pluginReportStorage) getNetDNSPluginReports(ctx context.Context, fileNameList []string) ([]*v1beta1.PluginReport, error) {
	var resList []*v1beta1.PluginReport

	netDNSList, err := p.clientSet.SpiderdoctorV1beta1().Netdnses().List(ctx, metav1.ListOptions{})
	if nil != err {
		return nil, err
	}

	netDNSFileNameList := func() []string {
		var arr []string
		for _, fileName := range fileNameList {
			if strings.HasPrefix(fileName, v1beta1.NetDNSTaskName) {
				if strings.Contains(fileName, summary) {
					continue
				}
				arr = append(arr, fileName)
			}
		}
		sort.Strings(arr)
		return arr
	}()

	for _, netDNS := range netDNSList.Items {
		tmpNetDNS := netDNS.DeepCopy()
		if tmpNetDNS.Status.DoneRound == nil || tmpNetDNS.Status.ExpectedRound == nil {
			klog.Infof("NetDNS %s has no expectedRound or no done round", tmpNetDNS.Name)
			continue
		}

		result, latestRoundNumber, err := p.getLatestRoundReports(tmpNetDNS.Name, netDNSFileNameList)
		if nil != err {
			return nil, err
		}

		// TODO (Icarus9913): redesign this
		var taskStatus string
		if tmpNetDNS.Status.Finish {
			taskStatus = "Finished"
		} else {
			taskStatus = "NotFinished"
		}

		var finishedRoundNumber int64
		if len(tmpNetDNS.Status.History) != 0 {
			finishedRoundNumber = int64(tmpNetDNS.Status.History[0].RoundNumber)
		}

		pluginReportSpec := v1beta1.PluginReportSpec{
			TaskName:            tmpNetDNS.Name,
			TaskType:            v1beta1.NetDNSTaskName,
			ToTalRoundNumber:    *tmpNetDNS.Status.ExpectedRound,
			FinishedRoundNumber: finishedRoundNumber,
			FailedRoundNumber:   nil,
			Status:              taskStatus,
			ReportRoundNumber:   latestRoundNumber,
			Report:              result,
		}

		pluginReport := &v1beta1.PluginReport{}
		pluginReport.Name = tmpNetDNS.Name
		pluginReport.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1beta1.GroupName,
			Version: v1beta1.V1betaVersion,
			Kind:    v1beta1.KindPluginReport,
		})
		pluginReport.Spec = pluginReportSpec

		resList = append(resList, pluginReport)
	}

	return resList, nil
}

func (p pluginReportStorage) getHttpAppHealthyReports(ctx context.Context, fileNameList []string) ([]*v1beta1.PluginReport, error) {
	var resList []*v1beta1.PluginReport

	httpAppHealthyList, err := p.clientSet.SpiderdoctorV1beta1().HttpAppHealthies().List(ctx, metav1.ListOptions{})
	if nil != err {
		return nil, err
	}

	httpAppHealthyNameList := func() []string {
		var arr []string
		for _, fileName := range fileNameList {
			if strings.Contains(fileName, v1beta1.HttpAppHealthyTaskName) {
				if strings.Contains(fileName, summary) {
					continue
				}
				arr = append(arr, fileName)
			}
		}
		sort.Strings(arr)
		return arr
	}()

	for _, httpAppHealthy := range httpAppHealthyList.Items {
		tmpHttpAppHealthy := httpAppHealthy.DeepCopy()
		if httpAppHealthy.Status.DoneRound == nil || httpAppHealthy.Status.ExpectedRound == nil {
			klog.Infof("HttpAppHealthy %s has no expectedRound or no done round", tmpHttpAppHealthy.Name)
			continue
		}

		result, latestRoundNumber, err := p.getLatestRoundReports(tmpHttpAppHealthy.Name, httpAppHealthyNameList)
		if nil != err {
			return nil, err
		}

		// TODO (Icarus9913): redesign this
		var taskStatus string
		if tmpHttpAppHealthy.Status.Finish {
			taskStatus = "Finished"
		} else {
			taskStatus = "NotFinished"
		}

		var finishedRoundNumber int64
		if len(tmpHttpAppHealthy.Status.History) != 0 {
			finishedRoundNumber = int64(tmpHttpAppHealthy.Status.History[0].RoundNumber)
		}

		pluginReportSpec := v1beta1.PluginReportSpec{
			TaskName:            tmpHttpAppHealthy.Name,
			TaskType:            v1beta1.HttpAppHealthyTaskName,
			ToTalRoundNumber:    *tmpHttpAppHealthy.Status.ExpectedRound,
			FinishedRoundNumber: finishedRoundNumber,
			FailedRoundNumber:   nil,
			Status:              taskStatus,
			ReportRoundNumber:   latestRoundNumber,
			Report:              result,
		}

		pluginReport := &v1beta1.PluginReport{}
		pluginReport.Name = tmpHttpAppHealthy.Name
		pluginReport.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1beta1.GroupName,
			Version: v1beta1.V1betaVersion,
			Kind:    v1beta1.KindPluginReport,
		})
		pluginReport.Spec = pluginReportSpec

		resList = append(resList, pluginReport)
	}

	return resList, nil
}

func (p pluginReportStorage) getNetReachHealthyReports(ctx context.Context, fileNameList []string) ([]*v1beta1.PluginReport, error) {
	var resList []*v1beta1.PluginReport

	netReachHealthyList, err := p.clientSet.SpiderdoctorV1beta1().NetReachHealthies().List(ctx, metav1.ListOptions{})
	if nil != err {
		return nil, err
	}

	netReachHealthyFileNameList := func() []string {
		var arr []string
		for _, fileName := range fileNameList {
			if strings.HasPrefix(fileName, v1beta1.NetReachHealthyTaskName) {
				if strings.Contains(fileName, summary) {
					continue
				}
				arr = append(arr, fileName)
			}
		}
		sort.Strings(arr)
		return arr
	}()

	for _, netReachHealthy := range netReachHealthyList.Items {
		tmpNetReachHealthy := netReachHealthy.DeepCopy()
		if tmpNetReachHealthy.Status.DoneRound == nil || tmpNetReachHealthy.Status.ExpectedRound == nil {
			klog.Infof("NetReachHealthy %s has no expectedRound or no done round", tmpNetReachHealthy.Name)
			continue
		}

		result, latestRoundNumber, err := p.getLatestRoundReports(tmpNetReachHealthy.Name, netReachHealthyFileNameList)
		if nil != err {
			return nil, err
		}

		// TODO (Icarus9913): redesign this
		var taskStatus string
		if tmpNetReachHealthy.Status.Finish {
			taskStatus = "Finished"
		} else {
			taskStatus = "NotFinished"
		}

		var finishedRoundNumber int64
		if len(tmpNetReachHealthy.Status.History) != 0 {
			finishedRoundNumber = int64(tmpNetReachHealthy.Status.History[0].RoundNumber)
		}

		pluginReportSpec := v1beta1.PluginReportSpec{
			TaskName:            tmpNetReachHealthy.Name,
			TaskType:            v1beta1.NetReachHealthyTaskName,
			ToTalRoundNumber:    *tmpNetReachHealthy.Status.ExpectedRound,
			FinishedRoundNumber: finishedRoundNumber,
			FailedRoundNumber:   nil,
			Status:              taskStatus,
			ReportRoundNumber:   latestRoundNumber,
			Report:              result,
		}

		pluginReport := &v1beta1.PluginReport{}
		pluginReport.Name = tmpNetReachHealthy.Name
		pluginReport.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
			Group:   v1beta1.GroupName,
			Version: v1beta1.V1betaVersion,
			Kind:    v1beta1.KindPluginReport,
		})
		pluginReport.Spec = pluginReportSpec

		resList = append(resList, pluginReport)
	}

	return resList, nil
}
