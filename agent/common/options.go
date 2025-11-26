package common

import (
	"container/list"
	"context"
	anc "claudeinsight/agent/analysis/common"
	"claudeinsight/agent/compatible"
	"claudeinsight/agent/protocol"
	"claudeinsight/agent/render/watch"
	"claudeinsight/bpf"
	"claudeinsight/common"
	"os"
	"runtime"
)

type LoadBpfProgramFunction func() *list.List
type InitCompletedHook func()
type ConnManagerInitHook func(any)

var Options *AgentOptions

const perfEventDataBufferSize = 30 * 1024 * 1024
const perfEventControlBufferSize = 1 * 1024 * 1024

type AgentOptions struct {
	Stopper                chan os.Signal
	CustomSyscallEventHook bpf.SyscallEventHook
	CustomConnEventHook    bpf.ConnEventHook
	CustomKernEventHook    bpf.KernEventHook
	CustomSslEventHook     bpf.SslEventHook
	InitCompletedHook      InitCompletedHook
	ConnManagerInitHook    ConnManagerInitHook
	LoadBpfProgramFunction LoadBpfProgramFunction
	ProcessorsNum          int
	MessageFilter          protocol.ProtocolFilter
	LatencyFilter          protocol.LatencyFilter
	TraceSide              common.SideEnum
	IfName                 string
	BTFFilePath            string
	protocol.SizeFilter
	AnalysisEnable bool
	anc.AnalysisOptions
	PerfEventBufferSizeForData  int
	PerfEventBufferSizeForEvent int
	WatchOptions                watch.WatchOptions
	PerformanceMode             bool
	ConntrackCloseWaitTimeMills int
	MaxAllowStuckTimeMills      int
	StartGopsServer             bool

	FilterComm              string
	ProcessExecEventChannel chan *bpf.AgentProcessExecEvent

	Objs                any
	Ctx                 context.Context
	Kv                  *compatible.KernelVersion
	LoadPorgressChannel chan string

	SyscallPerfEventMapPageNum int
	SslPerfEventMapPageNum     int
	ConnPerfEventMapPageNum    int
	KernPerfEventMapPageNum    int
	FirstPacketEventMapPageNum int
}

func ValidateAndRepairOptions(options AgentOptions) AgentOptions {
	var newOptions = options
	if newOptions.Stopper == nil {
		newOptions.Stopper = make(chan os.Signal)
	}
	if newOptions.ProcessorsNum == 0 {
		newOptions.ProcessorsNum = runtime.NumCPU()
	}
	if newOptions.MessageFilter == nil {
		newOptions.MessageFilter = protocol.BaseFilter{}
	}
	if newOptions.PerfEventBufferSizeForData <= 0 {
		newOptions.PerfEventBufferSizeForData = perfEventDataBufferSize
	}
	if newOptions.PerfEventBufferSizeForEvent <= 0 {
		newOptions.PerfEventBufferSizeForEvent = perfEventControlBufferSize
	}
	newOptions.WatchOptions.Init()
	newOptions.LoadPorgressChannel = make(chan string, 10)
	return newOptions
}
