package pluginManager

import (
	"context"
	"fmt"
	plugintypes "github.com/spidernet-io/spiderdoctor/pkg/pluginManager/types"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

type pluginAgentReconciler struct {
	client client.Client
	p      plugintypes.ChainingPlugin
	logger *zap.Logger
}

func (s *pluginAgentReconciler) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	return s.p.AgentReconcile(s.logger, s.client, ctx, r)
}

var _ reconcile.Reconciler = &pluginAgentReconciler{}

func (s *pluginManager) runAgentReconcile() {
	logger := s.logger

	scheme := runtime.NewScheme()
	for name, plugin := range s.chainingPlugins {
		if e := plugin.AddToScheme(scheme); e != nil {
			logger.Sugar().Fatalf("failed to add scheme for plugin, reason=%v", name, e)
		}
	}
	n := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), n)
	if err != nil {
		logger.Sugar().Fatalf("failed to NewManager, reason=%v", err)
	}

	for name, plugin := range s.chainingPlugins {
		builder := ctrl.NewControllerManagedBy(mgr)
		logger.Sugar().Infof("run controller for plugin %v", name)
		k := &pluginAgentReconciler{
			logger: logger.Named(name + "Reconciler"),
			p:      plugin,
		}
		b, e := builder.For(plugin.GetApiType()).Owns(plugin.GetApiType()).Build(k)
		if e != nil {
			s.logger.Sugar().Fatalf("failed to builder reconcile for plugin %v, error=%v", name, e)
		}
		if e := b.Watch(&source.Kind{Type: plugin.GetApiType()}, &handler.EnqueueRequestForObject{}); e != nil {
			s.logger.Sugar().Fatalf("failed to watch for plugin %v, error=%v", name, e)
		}
		go func(name string) {
			msg := fmt.Sprintf("reconcile of plugin %v down", name)
			if e := b.Start(context.Background()); e != nil {
				msg += fmt.Sprintf(", error=%v", e)
			}
			s.logger.Error(msg)
			time.Sleep(5 * time.Second)
		}(name)
	}

	go func() {
		msg := fmt.Sprintf("reconcile of plugin down")
		if e := mgr.Start(context.Background()); e != nil {
			msg += fmt.Sprintf(", error=%v", e)
		}
		s.logger.Error(msg)
		time.Sleep(5 * time.Second)
	}()
}
