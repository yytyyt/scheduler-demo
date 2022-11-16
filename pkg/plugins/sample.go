package plugins

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

const (
	//Name 定义插件名称
	Name              = "sample-plugin"
	preFilterStateKey = "Prefilter" + Name
)

var _ framework.PreFilterPlugin = &Sample{}
var _ framework.FilterPlugin = &Sample{}

type SampleArgs struct {
	FavoriteColor  string `json:"favorColor,omitempty"`
	FavoriteNumber int    `json:"favoriteNumber,omitempty"`
	ThanksTo       string `json:"thanksTo,omitempty"`
}

// 获取插件配置得参数
func getSampleArgs(object runtime.Object) (*SampleArgs, error) {
	sa := &SampleArgs{}
	err := frameworkruntime.DecodeInto(object, sa)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

type preFilterState struct {
	framework.Resource
}

func (s *preFilterState) Clone() framework.StateData {
	return s
}

func getPreFilterState(state *framework.CycleState) (*preFilterState, error) {
	data, err := state.Read(preFilterStateKey)
	if err != nil {
		return nil, err
	}
	filterState, ok := data.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v convert to SamplePlugin preFilterState error", filterState)
	}
	return filterState, nil
}

type Sample struct {
	args   *SampleArgs
	handle framework.FrameworkHandle
}

func (s *Sample) Name() string {
	return Name
}

func computePodResourceLimit(pod *corev1.Pod) *preFilterState {
	result := &preFilterState{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Limits)
	}
	return result
}

func (s *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod) *framework.Status {
	klog.V(3).Infof("prefilter pod:%v", pod.Name)
	state.Write(preFilterStateKey, computePodResourceLimit(pod))
	return nil
}

func (s *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	filterState, err := getPreFilterState(state)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	// 业务处理
	klog.V(3).Infof("filter pod:%v,node:%v,pre state:%v", pod.Name, nodeInfo.Node().Name, filterState)
	return framework.NewStatus(framework.Success, "")
}

//func(configuration runtime.Object, f v1alpha1.FrameworkHandle) (v1alpha1.Plugin, error)

func New(object runtime.Object, f framework.FrameworkHandle) (framework.Plugin, error) {
	args, err := getSampleArgs(object)
	if err != nil {
		return nil, err
	}
	klog.V(3).Infof("get plugin config args:%+v", args)
	return &Sample{
		args:   args,
		handle: f,
	}, nil
}
