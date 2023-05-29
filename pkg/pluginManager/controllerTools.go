// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package pluginManager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	k8sObjManager "github.com/spidernet-io/spiderdoctor/pkg/k8ObjManager"
	crd "github.com/spidernet-io/spiderdoctor/pkg/k8s/apis/spiderdoctor.spidernet.io/v1beta1"
	plugintypes "github.com/spidernet-io/spiderdoctor/pkg/pluginManager/types"
	"github.com/spidernet-io/spiderdoctor/pkg/reportManager"
	"github.com/spidernet-io/spiderdoctor/pkg/types"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

func (s *pluginControllerReconciler) GetSpiderAgentNodeNotInRecord(ctx context.Context, succeedNodeList []string, agentNodeSelector *metav1.LabelSelector) (failNodelist []string, err error) {
	var allNodeList []string
	var e error
	if agentNodeSelector == nil {
		allNodeList, e = k8sObjManager.GetK8sObjManager().ListDaemonsetPodNodes(ctx, types.ControllerConfig.Configmap.SpiderDoctorAgentDaemonsetName, types.ControllerConfig.PodNamespace)
		if e != nil {
			return nil, e
		}
		s.logger.Sugar().Debugf("all agent nodes: %+v", allNodeList)
		if len(allNodeList) == 0 {
			return nil, fmt.Errorf("failed to find agent node ")
		}
	} else {
		allNodeList, e = k8sObjManager.GetK8sObjManager().ListSelectedNodes(ctx, agentNodeSelector)
		if e != nil {
			return nil, e
		}
		s.logger.Sugar().Debugf("selected agent nodes: %+v", allNodeList)
		if len(allNodeList) == 0 {
			return nil, fmt.Errorf("failed to find agent node ")
		}
	}

	failNodelist = []string{}
OUTER:
	for _, v := range allNodeList {
		for _, m := range succeedNodeList {
			if m == v {
				continue OUTER
			}
		}
		failNodelist = append(failNodelist, v)
	}
	return failNodelist, nil
}

func (s *pluginControllerReconciler) UpdateRoundFinalStatus(logger *zap.Logger, ctx context.Context, newStatus *crd.TaskStatus, agentNodeSelector *metav1.LabelSelector, deadline bool) (roundDone bool, err error) {

	latestRecord := &(newStatus.History[0])
	roundNumber := latestRecord.RoundNumber

	if latestRecord.Status == crd.StatusHistoryRecordStatusFail || latestRecord.Status == crd.StatusHistoryRecordStatusSucceed || latestRecord.Status == crd.StatusHistoryRecordStatusNotstarted {
		return true, nil
	}

	// when not reach deadline, ignore when nothing report
	if !deadline && len(latestRecord.SucceedAgentNodeList) == 0 && len(latestRecord.FailedAgentNodeList) == 0 {
		logger.Sugar().Debugf("round %v not report anything", roundNumber)
		return false, nil
	}

	// update result in latestRecord
	reportNode := []string{}
	reportNode = append(reportNode, latestRecord.SucceedAgentNodeList...)
	reportNode = append(reportNode, latestRecord.FailedAgentNodeList...)
	if unknowReportNodeList, e := s.GetSpiderAgentNodeNotInRecord(ctx, reportNode, agentNodeSelector); e != nil {
		logger.Sugar().Errorf("round %v failed to GetSpiderAgentNodeNotInSucceedRecord, error=%v", roundNumber, e)
		return false, e
	} else {
		if len(unknowReportNodeList) > 0 && !deadline {
			// when not reach the deadline, ignore
			logger.Sugar().Debugf("round %v , partial agents did not reported, wait for daedline", roundNumber)
			return false, nil
		}

		// it's ok to collect round status
		if len(unknowReportNodeList) > 0 || len(latestRecord.FailedAgentNodeList) > 0 {
			latestRecord.NotReportAgentNodeList = unknowReportNodeList
			n := crd.StatusHistoryRecordStatusFail
			latestRecord.Status = n
			newStatus.LastRoundStatus = &n
			logger.Sugar().Errorf("round %v failed , failedNode=%v, unknowReportNode=%v", roundNumber, latestRecord.FailedAgentNodeList, unknowReportNodeList)

			if len(latestRecord.FailedAgentNodeList) > 0 {
				latestRecord.FailureReason = "some agents failed"
			} else if len(unknowReportNodeList) > 0 {
				latestRecord.FailureReason = "some agents did not report"
			}

		} else {
			n := crd.StatusHistoryRecordStatusSucceed
			latestRecord.Status = n
			newStatus.LastRoundStatus = &n
			logger.Sugar().Infof("round %v succeeded ", latestRecord.RoundNumber)
		}
		cnt := len(reportNode) + len(unknowReportNodeList)
		latestRecord.ExpectedActorNumber = &cnt
		latestRecord.EndTimeStamp = &metav1.Time{
			Time: time.Now(),
		}
		i := time.Since(latestRecord.StartTimeStamp.Time).String()
		latestRecord.Duration = &i

		return true, nil
	}

}

func (s *pluginControllerReconciler) WriteSummaryReport(taskName string, roundNumber int, newStatus *crd.TaskStatus) {
	if s.fm == nil {
		return
	}

	kindName := strings.Split(taskName, ".")[0]
	instanceName := strings.TrimPrefix(taskName, kindName+".")
	t := time.Duration(types.ControllerConfig.ReportAgeInDay*24) * time.Hour
	endTime := newStatus.History[0].StartTimeStamp.Add(t)

	if !s.fm.CheckTaskFileExisted(kindName, instanceName, roundNumber) {
		// add to workqueue to collect all report of last round, for node latestRecord.FailedAgentNodeList and latestRecord.SucceedAgentNodeList
		reportManager.TriggerSyncReport(fmt.Sprintf("%s.%d", taskName, roundNumber))

		// TODO (Icarus9913): change to use v1beta1.Report ?
		// write controller summary report
		msg := plugintypes.PluginReport{
			TaskName:       strings.ToLower(taskName),
			TaskSpec:       "",
			RoundNumber:    roundNumber,
			RoundResult:    plugintypes.RoundResultStatus(newStatus.History[0].Status),
			FailedReason:   newStatus.History[0].FailureReason,
			NodeName:       "",
			PodName:        types.ControllerConfig.PodName,
			StartTimeStamp: newStatus.History[0].StartTimeStamp.Time,
			EndTimeStamp:   time.Now(),
			RoundDuraiton:  time.Since(newStatus.History[0].StartTimeStamp.Time).String(),
			Detail:         newStatus.History[0],
			ReportType:     plugintypes.ReportTypeSummary,
		}

		if jsongByte, e := json.Marshal(msg); e != nil {
			s.logger.Sugar().Errorf("failed to generate round summary report for kind %v task %v round %v, json marsha error=%v", kindName, instanceName, roundNumber, e)
		} else {
			// print to stdout for human reading
			fmt.Printf("%+v\n ", string(jsongByte))

			var out bytes.Buffer
			if e := json.Indent(&out, jsongByte, "", "\t"); e != nil {
				s.logger.Sugar().Errorf("failed to generate round summary report for kind %v task %v round %v, json Indent error=%v", kindName, instanceName, roundNumber, e)
			} else {
				// file name format: fmt.Sprintf("%s_%s_round%d_%s_%s", kindName, taskName, roundNumber, nodeName, suffix)
				if e := s.fm.WriteTaskFile(kindName, instanceName, roundNumber, "summary", endTime, out.Bytes()); e != nil {
					s.logger.Sugar().Errorf("failed to generate round summary report for kind %v task %v round %v, write file error=%v", kindName, instanceName, roundNumber, e)
				} else {
					s.logger.Sugar().Debugf("succeeded to generate round summary report for kind %v task %v round %v", kindName, instanceName, roundNumber)
				}
			}
		}
	}
}

func (s *pluginControllerReconciler) UpdateStatus(logger *zap.Logger, ctx context.Context, oldStatus *crd.TaskStatus, schedulePlan *crd.SchedulePlan, sourceAgent *metav1.LabelSelector, taskName string) (result *reconcile.Result, taskStatus *crd.TaskStatus, e error) {
	newStatus := oldStatus.DeepCopy()
	nextInterval := time.Duration(types.ControllerConfig.Configmap.TaskPollIntervalInSecond) * time.Second
	nowTime := time.Now()
	var startTime time.Time
	// init new instance first
	scheduler := NewSchedule(*schedulePlan.Schedule)
	if newStatus.ExpectedRound == nil || len(newStatus.History) == 0 {
		startTime = scheduler.StartTime(nowTime)
		m := int64(0)
		newStatus.DoneRound = &m
		newStatus.ExpectedRound = &schedulePlan.RoundNumber

		newRecord := NewStatusHistoryRecord(startTime, 1, schedulePlan)
		newStatus.History = append(newStatus.History, *newRecord)
		logger.Sugar().Debugf("initialize the first round of task : %v ", taskName, *newRecord)
		// trigger
		result = &reconcile.Result{
			Requeue: true,
		}
		// updating status firstly , it will trigger to handle it next round
		return result, newStatus, nil
	}

	if *newStatus.DoneRound == *newStatus.ExpectedRound {
		// done task
		return nil, nil, nil
	}

	latestRecord := &(newStatus.History[0])
	roundNumber := latestRecord.RoundNumber
	logger.Sugar().Debugf("current time:%v , latest history record: %+v", nowTime, latestRecord)
	logger.Sugar().Debugf("all history record: %+v", newStatus.History)

	switch {
	case nowTime.After(latestRecord.StartTimeStamp.Time) && nowTime.Before(latestRecord.DeadLineTimeStamp.Time):

		if latestRecord.Status == crd.StatusHistoryRecordStatusNotstarted {
			latestRecord.Status = crd.StatusHistoryRecordStatusOngoing
			// requeue immediately to make sure the update succeed , not conflicted
			result = &reconcile.Result{
				Requeue: true,
			}

		} else if latestRecord.Status == crd.StatusHistoryRecordStatusOngoing {
			logger.Debug("try to poll the status of task " + taskName)
			if roundDone, e := s.UpdateRoundFinalStatus(logger, ctx, newStatus, sourceAgent, false); e != nil {
				return nil, nil, e
			} else {
				if roundDone {
					logger.Sugar().Infof("round %v get reports from all agents ", roundNumber)

					// before insert new record, write summary of last round
					s.WriteSummaryReport(taskName, roundNumber, newStatus)

					// add new round record
					if *(newStatus.DoneRound) < *(newStatus.ExpectedRound) || *newStatus.ExpectedRound == -1 {
						n := *(newStatus.DoneRound) + 1
						newStatus.DoneRound = &n
						startTime = scheduler.Next(latestRecord.StartTimeStamp.Time)
						if n < *(newStatus.ExpectedRound) || *newStatus.ExpectedRound == -1 {

							newRecord := NewStatusHistoryRecord(startTime, int(n+1), schedulePlan)

							tmp := append([]crd.StatusHistoryRecord{*newRecord}, newStatus.History...)
							if len(tmp) > types.ControllerConfig.Configmap.CrdMaxHistory {
								tmp = tmp[:(types.ControllerConfig.Configmap.CrdMaxHistory)]
							}
							newStatus.History = tmp

							logger.Sugar().Infof("insert new record for next round : %+v", *newRecord)
						} else {
							newStatus.Finish = true
						}
					}

					// requeue immediately to make sure the update succeed , not conflicted
					result = &reconcile.Result{
						Requeue: true,
					}
				} else {
					// trigger after interval
					result = &reconcile.Result{
						RequeueAfter: nextInterval,
					}
				}
			}
		} else {
			logger.Debug("ignore poll the finished round of task " + taskName)

			// trigger when deadline
			result = &reconcile.Result{
				RequeueAfter: time.Until(latestRecord.DeadLineTimeStamp.Time),
			}
		}

	case nowTime.Before(latestRecord.StartTimeStamp.Time):
		fallthrough
	case nowTime.After(latestRecord.DeadLineTimeStamp.Time):
		if *newStatus.DoneRound == *newStatus.ExpectedRound {
			logger.Sugar().Debugf("task %s finish, ignore ", taskName)
			newStatus.Finish = true
			result = nil

		} else {

			// when task not finish , once we update the status succeed , we will not get here , it should go to case nowTime.Before(latestRecord.StartTimeStamp.Time)
			if latestRecord.Status == crd.StatusHistoryRecordStatusOngoing {
				// here, we should update last round status

				if _, e := s.UpdateRoundFinalStatus(logger, ctx, newStatus, sourceAgent, true); e != nil {
					return nil, nil, e
				} else {
					// all agent finished, so try to update the summary
					logger.Sugar().Infof("round %v got reports from all agents, try to summarize", roundNumber)

					// before insert new record, write summary of last round
					s.WriteSummaryReport(taskName, roundNumber, newStatus)

					// add new round record
					if *(newStatus.DoneRound) < *(newStatus.ExpectedRound) || *newStatus.ExpectedRound == -1 {
						n := *(newStatus.DoneRound) + 1
						newStatus.DoneRound = &n

						if n < *(newStatus.ExpectedRound) || *newStatus.ExpectedRound == -1 {
							startTime = scheduler.Next(latestRecord.StartTimeStamp.Time)
							newRecord := NewStatusHistoryRecord(startTime, int(n+1), schedulePlan)
							tmp := append([]crd.StatusHistoryRecord{*newRecord}, newStatus.History...)
							if len(tmp) > types.ControllerConfig.Configmap.CrdMaxHistory {
								tmp = tmp[:(types.ControllerConfig.Configmap.CrdMaxHistory)]
							}
							newStatus.History = tmp

							logger.Sugar().Infof("insert new record for next round : %+v", *newRecord)
						} else {
							newStatus.Finish = true
						}
					}

					// requeue immediately to make sure the update succeed , not conflicted
					result = &reconcile.Result{
						Requeue: true,
					}
				}
			} else {
				// round finish
				// trigger when next round start
				currentLatestRecord := &(newStatus.History[0])
				logger.Sugar().Infof("task %v wait for next round %v at %v", taskName, currentLatestRecord.RoundNumber, currentLatestRecord.StartTimeStamp)
				result = &reconcile.Result{
					RequeueAfter: time.Until(currentLatestRecord.StartTimeStamp.Time),
				}
			}
		}
	}

	return result, newStatus, nil

}
