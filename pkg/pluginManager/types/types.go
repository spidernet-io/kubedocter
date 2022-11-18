// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type ChainingPlugin interface {
	GetApiType() client.Object

	AgentEexecuteTask(logger *zap.Logger, ctx context.Context, obj runtime.Object) (failureReason string, report PluginRoundDetail, err error)

	// ControllerReconcile(*zap.Logger, client.Client, context.Context, reconcile.Request) (reconcile.Result, error)
	// AgentReconcile(*zap.Logger, client.Client, context.Context, reconcile.Request) (reconcile.Result, error)

	WebhookMutating(logger *zap.Logger, ctx context.Context, obj runtime.Object) error
	WebhookValidateCreate(logger *zap.Logger, ctx context.Context, obj runtime.Object) error
	WebhookValidateUpdate(logger *zap.Logger, ctx context.Context, oldObj, newObj runtime.Object) error
}

type RoundResultStatus string

const (
	RoundResultSucceed = RoundResultStatus("succeed")
	RoundResultFail    = RoundResultStatus("fail")
)

type PluginReport struct {
	TaskName       string
	TaskSpec       interface{}
	RoundNumber    int
	RoundResult    RoundResultStatus
	AgentNodeName  string
	AgentPodName   string
	FailedReason   string
	StartTimeStamp time.Time
	EndTimeStamp   time.Time
	RoundDuraiton  string
	Detail         PluginRoundDetail
}

type PluginRoundDetail map[string]interface{}

const (
	ApiMsgGetFailure      = "failed to get instance"
	ApiMsgUnknowCRD       = "unsupported crd type"
	ApiMsgUnsupportModify = "unsupported modify spec"
)
