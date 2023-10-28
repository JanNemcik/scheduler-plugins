package v2xcontext

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/apis/config"
	"sigs.k8s.io/scheduler-plugins/pkg/networktraffic"
)

// ContextPrio is a queue sort plugin that favors pods based on their
// contextual priority. Pods with higher priority are favored.
// Implements framework.QueueSortPlugin
type ContextPrio struct {
	handle     framework.Handle
	prometheus *networktraffic.PrometheusHandle
}

// Name is the name of the plugin used in the Registry and configurations.
const Name = "ContextPrio"

var _ = framework.QueueSortPlugin(&ContextPrio{})

// Name returns name of the plugin. It is used in logs, etc.
func (cp *ContextPrio) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args, ok := obj.(*config.ContextPrioArgs)
	if !ok {
		return nil, fmt.Errorf("[ContextPrio] want args to be of type ContextPrioArgs, got %T", obj)
	}

	klog.Infof("[ContextPrio] args received. NetworkInterface: %s; TimeRangeInMinutes: %d, Address: %s", args.NetworkInterface, args.TimeRangeInMinutes, args.Address)

	return &ContextPrio{
		handle:     h,
		prometheus: networktraffic.NewPrometheus(args.Address, args.NetworkInterface, time.Minute*time.Duration(args.TimeRangeInMinutes)),
	}, nil
}

func (cp *ContextPrio) Less(first *framework.QueuedPodInfo, second *framework.QueuedPodInfo) bool {
	s_ctx, ok := second.Pod.Annotations["v2x.context"]
	if !ok {
		klog.Infof("[ContextPrio] context for comparison not present")
		return true
	}
	f_ctx, f_present := first.Pod.Annotations["v2x.context"]
	if !f_present {
		klog.Infof("[ContextPrio] context for comparison not present")
		return true
	}
	klog.Infof("[ContextPrio] contexts identified. Contexts: %s, %s", f_ctx, s_ctx)
	return true
}
