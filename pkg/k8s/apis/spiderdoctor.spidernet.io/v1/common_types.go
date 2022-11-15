// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SchedulePlan struct {
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	StartAfterMinute int64 `json:"startAfterMinute"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	RoundNumber int64 `json:"roundNumber"`

	// +kubebuilder:default=360
	// +kubebuilder:validation:Minimum=1
	IntervalMinute int64 `json:"intervalMinute"`

	// +kubebuilder:default=60
	// +kubebuilder:validation:Minimum=1
	TimeoutMinute int64 `json:"timeoutMinute"`
}

type TaskStatus struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	ExpectedRound *int64 `json:"expectedRound,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	DoneRound *int64 `json:"doneRound,omitempty"`

	Finish bool `json:"finish"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type:=string
	// +kubebuilder:validation:Format:=date-time
	LastRoundFinishTimeStamp *metav1.Time `json:"lastRoundFinishTimeStamp,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=succeed;fail;unknown
	LastRoundStatus *string `json:"lastRoundStatus,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type:=string
	// +kubebuilder:validation:Format:=date-time
	NextRoundStartTimeStamp *metav1.Time `json:"nextRoundStartTimeStamp,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type:=string
	// +kubebuilder:validation:Format:=date-time
	NextRoundDeadLineTimeStamp *metav1.Time `json:"nextRoundDeadLineTimeStamp,omitempty"`

	History []StatusHistoryRecord `json:"history"`
}

type StatusHistoryRecord struct {

	// +kubebuilder:validation:Enum=succeed;fail;unknown
	Status string `json:"status"`

	// +kubebuilder:validation:Type:=string
	// +kubebuilder:validation:Format:=date-time
	StartTimeStamp *metav1.Time `json:"startTimeStamp,omitempty"`

	FailedAgentNodeList []string `json:"failedAgentNodeList"`
}

type NetSuccessCondition struct {

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	// +kubebuilder:validation:Optional
	SuccessRate *float32 `json:"successRate,omitempty"`

	// +kubebuilder:default=5000
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Optional
	MeanAccessDelayInMs *uint64 `json:"meanAccessDelayInMs,omitempty"`
}
