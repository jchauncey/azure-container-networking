// Copyright Microsoft. All rights reserved.
package logger

import (
	"fmt"

	"github.com/Azure/azure-container-networking/aitelemetry"
	"github.com/Azure/azure-container-networking/log"
)

const (
	// Wait time for closing AI telemetry session.
	waitTimeInSecs = 10
)

var (
	Log        *CNSLogger
	aiMetadata string
)

type CNSLogger struct {
	logger               *log.Logger
	th                   aitelemetry.TelemetryHandle
	Orchestrator         string
	NodeID               string
	DisableTraceLogging  bool
	DisableMetricLogging bool
}

// Initialize CNS Logger
func InitLogger(fileName string, logLevel, logTarget int, logDir string) {
	Log = &CNSLogger{
		logger: log.NewLogger(fileName, logLevel, logTarget, logDir),
	}
}

// Intialize CNS AI telmetry instance
func InitAI(aiConfig aitelemetry.AIConfig, disableTraceLogging, disableMetricLogging bool) {
	var err error

	Log.th, err = aitelemetry.NewAITelemetry("", aiMetadata, aiConfig)
	if err != nil {
		Log.logger.Errorf("Error initializing AI Telemetry:%v", err)
		return
	}

	Log.logger.Printf("AI Telemetry Handle created")
	Log.DisableMetricLogging = disableMetricLogging
	Log.DisableTraceLogging = disableTraceLogging
}

func InitReportChannel(reports chan interface{}) {
	Log.logger.SetChannel(reports)
}

// Close CNS and AI telemetry handle
func Close() {
	Log.logger.Close()
	if Log.th != nil {
		Log.th.Close(waitTimeInSecs)
	}
}

func SetTargetLogDirectory(target int, dir string) error {
	return Log.logger.SetTargetLogDirectory(target, dir)
}

// Set context details for logs and metrics
func SetContextDetails(orchestrator string, nodeID string) {
	Printf("SetContext details called with: %v orchestrator nodeID %v", orchestrator, nodeID)
	Log.Orchestrator = orchestrator
	Log.NodeID = nodeID
}

// Send AI telemetry trace
func sendTraceInternal(msg string) {
	report := aitelemetry.Report{CustomDimensions: make(map[string]string)}
	report.Message = msg
	report.CustomDimensions[OrchestratorTypeStr] = Log.Orchestrator
	report.CustomDimensions[NodeIDStr] = Log.NodeID
	report.Context = Log.NodeID
	Log.th.TrackLog(report)
}

func Printf(format string, args ...interface{}) {
	Log.logger.Printf(format, args...)

	if Log.th == nil || Log.DisableTraceLogging {
		return
	}

	msg := fmt.Sprintf(format, args...)
	sendTraceInternal(msg)
}

func Debugf(format string, args ...interface{}) {
	Log.logger.Debugf(format, args...)

	if Log.th == nil || Log.DisableTraceLogging {
		return
	}

	msg := fmt.Sprintf(format, args...)
	sendTraceInternal(msg)
}

func Errorf(format string, args ...interface{}) {
	Log.logger.Errorf(format, args...)

	if Log.th == nil || Log.DisableTraceLogging {
		return
	}

	msg := fmt.Sprintf(format, args...)
	sendTraceInternal(msg)
}

func Request(tag string, request interface{}, err error) {
	Log.logger.Request(tag, request, err)
}

func Response(tag string, response interface{}, returnCode int, returnStr string, err error) {
	Log.logger.Response(tag, response, returnCode, returnStr, err)
}

// Send AI telemetry metric
func SendMetric(metric aitelemetry.Metric) {
	if Log.th == nil || Log.DisableMetricLogging {
		return
	}

	metric.CustomDimensions[OrchestratorTypeStr] = Log.Orchestrator
	metric.CustomDimensions[NodeIDStr] = Log.NodeID
	Log.th.TrackMetric(metric)
}
