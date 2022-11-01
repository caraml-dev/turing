package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/turing/api/turing/cluster"
	logger "github.com/caraml-dev/turing/api/turing/log"
)

const podLogTimeout = 10 * time.Second

// PodLogsV2 contains a list of logs in a kubernetes pod along with some extra information.
// This is the new format for pod logs
type PodLogsV2 struct {
	// Environment name of the router running the container that produces this log
	Environment string `json:"environment"`
	// Kubernetes namespace where the pod running the container is created
	Namespace string `json:"namespace"`
	// URL to dashboard since pods might be deleted after use
	// Since there are multiple pods, we will add the unique URLs here
	LoggingURL string `json:"logging_url"`
	// Logs is an array of logs, equivalent to one line of log
	Logs []*PodLogV2 `json:"logs"`
}

// PodLogV2 represents a single log line from a container running in Kubernetes.
type PodLogV2 struct {
	// Log timestamp in RFC3339 format
	Timestamp time.Time `json:"timestamp"`
	// Pod name running the container that produces this log
	PodName string `json:"pod_name"`
	// Log in text format, either TextPayload or JSONPayload will be set but not both
	TextPayload string `json:"text_payload,omitempty"`
}

// PodLog represents a single log line from a container running in Kubernetes.
// If the log message is in JSON format, JSONPayload must be populated with the
// structured JSON log message. Else, TextPayload must be populated with log text.
// This is the legacy format, use PodLogs for newer features instead.
type PodLog struct {
	// Log timestamp in RFC3339 format
	Timestamp time.Time `json:"timestamp"`
	// Environment name of the router running the container that produces this log
	Environment string `json:"environment"`
	// Kubernetes namespace where the pod running the container is created
	Namespace string `json:"namespace"`
	// Pod name running the container that produces this log
	PodName string `json:"pod_name"`
	// Container name that produces this log
	ContainerName string `json:"container_name,omitempty"`
	// Log in text format, either TextPayload or JSONPayload will be set but not both
	TextPayload string `json:"text_payload,omitempty"`
	// Log in JSON format, either TextPayload or JSONPayload will be set but not both
	JSONPayload map[string]interface{} `json:"json_payload,omitempty"`
}

// PodLogService is an interface to retrieve logs from Kubernetes Pods
type PodLogService interface {
	ListPodLogs(request PodLogRequest) ([]*PodLog, error)
}

type podLogService struct {
	clusterControllers map[string]cluster.Controller
}

// NewPodLogService creates a new PodLogService that deals with kubernetes pod logs
func NewPodLogService(clusterControllers map[string]cluster.Controller) PodLogService {
	return &podLogService{clusterControllers: clusterControllers}
}

// ConvertPodLogsToV2 converts to the new pod log format
func ConvertPodLogsToV2(namespace string, environment string, loggingURL string, podLogs []*PodLog) *PodLogsV2 {
	logs := &PodLogsV2{
		Environment: environment,
		Namespace:   namespace,
	}

	allLines := []*PodLogV2{}
	for _, p := range podLogs {
		line := &PodLogV2{
			Timestamp:   p.Timestamp,
			PodName:     p.PodName,
			TextPayload: p.TextPayload,
		}
		allLines = append(allLines, line)
	}

	logs.Logs = allLines
	logs.LoggingURL = loggingURL

	return logs
}

// PodLogRequest is the request for logs for a particular set of pods.
type PodLogRequest struct {
	// Kubernetes Namespace, usually the same as the project name
	Namespace string
	// Picks the logs from a selected container in a pod.
	DefaultContainer string
	// Environment that the pod is in, used for the cluster controller selection
	Environment string
	// Labels for Kubernetes pods
	LabelSelectors []LabelSelector
	// This is the template used for persistent logs that are outside Kubernetes
	LoggingURLTemplate *string
	// Container to get the logs from, default to 'user-container', the default container name in Knative
	Container string
	// If true, return the logs from previous terminated container in all pods
	Previous bool
	// (Optional) Timestamp from which to retrieve the logs from, useful for filtering recent logs. The logs retrieved
	// will have timestamp after (but not including) SinceTime.
	SinceTime *time.Time
	// (Optional) Number of lines from the end of the logs to retrieve. Should not be used together with HeadLines.
	// If both TailLines and Headlines are provided, TailLines will be ignored.
	TailLines *int64
	// (Optional) Number of lines from the start of the logs to retrieve.  Should not be used together with TailLines.
	// If both TailLines and Headlines are provided, TailLines will be ignored.
	HeadLines *int64
}

// LabelSelector refers to the label
type LabelSelector struct {
	Key   string
	Value string
}

func (s *LabelSelector) formatLabel() string {
	return fmt.Sprintf("%s=%s", s.Key, s.Value)
}

func formatLabelSelector(labels []LabelSelector) string {
	all := []string{}
	for _, l := range labels {
		all = append(all, l.formatLabel())
	}
	return strings.Join(all, ",")
}

func (s *podLogService) ListPodLogs(request PodLogRequest) ([]*PodLog, error) {
	// If both TailLines and Headlines are set, TailLines is ignored
	if request.TailLines != nil && request.HeadLines != nil {
		request.TailLines = nil
	}

	controller, ok := s.clusterControllers[request.Environment]
	if !ok {
		return nil, fmt.Errorf("cluster controller for environment '%s' not found", request.Environment)
	}

	labelSelector := formatLabelSelector(request.LabelSelectors)
	ctx, cancel := context.WithTimeout(context.Background(), podLogTimeout)
	defer cancel()

	pods, err := controller.ListPods(ctx, request.Namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	allPodLogs := make([]*PodLog, 0)
	for _, p := range pods.Items {
		logOpts := &v1.PodLogOptions{
			Container:  request.Container,
			Previous:   request.Previous,
			Timestamps: true,
		}

		// Only send tailLines argument to Kube API server if sinceTime is not set because Kube API will prioritize
		// tailLines over sinceTime and we want these options to be AND-ed together.
		// The log items returned by Kube API will later be trimmed separately according to the tailLines argument.
		if request.SinceTime == nil && request.TailLines != nil {
			logOpts.TailLines = request.TailLines
		}

		// Set default container if not set
		if logOpts.Container == "" {
			logOpts.Container = request.DefaultContainer
		}

		if request.SinceTime != nil {
			// Note that the requested sinceTime sent to Kubernetes API server is subtracted by 1 second
			// because Kubernetes API server only accepts sinceTime granularity to the second so we will
			// miss some logs within the second without this subtraction.
			logOpts.SinceTime = &metav1.Time{Time: request.SinceTime.Add(-time.Second)}
		}

		stream, err := controller.ListPodLogs(ctx, request.Namespace, p.Name, logOpts)
		if err != nil {
			// Error is handled here by logging it rather than returned because the caller usually does not know how to
			// handle it. Example of what can trigger ListPodLogs error: while the container is being created/terminated
			// Kubernetes API server will return error when logs are requested. In such case, it is better to return
			// empty logs and let the caller retry after the container becomes ready eventually.
			logger.Warnf("Failed to ListPodLogs: %s", err.Error())
			return allPodLogs, nil
		}

		scanner := bufio.NewScanner(stream)
		podLogs := make([]*PodLog, 0)
		for scanner.Scan() {
			logLine := scanner.Text()

			// A log line from Kubernetes API server will follow this format:
			// 2020-07-14T07:48:14.191189249Z {"msg":"log message"}
			timestampIndex := strings.Index(logLine, " ")
			if timestampIndex < 0 {
				// Missing expected RFC3339 timstamp in the log line, skip to next line
				continue
			}
			if (len(logLine) - 1) <= timestampIndex {
				// Empty log message, skip to next log line
				continue
			}

			timestamp, err := time.Parse(time.RFC3339, logLine[:timestampIndex])
			if err != nil {
				logger.Warnf("log message timestamp is not in RFC3339 format: %s", logLine[:timestampIndex])
				// Log timestamp value from Kube API server has invalid format, skip to next line
				continue
			}

			// We require this check because we send (SinceTime - 1sec) to Kube API Server
			if request.SinceTime != nil && (timestamp == *request.SinceTime || timestamp.Before(*request.SinceTime)) {
				continue
			}

			log := &PodLog{
				Timestamp:     timestamp,
				Environment:   request.Environment,
				Namespace:     request.Namespace,
				PodName:       p.Name,
				ContainerName: logOpts.Container,
			}

			logText := logLine[timestampIndex+1:]
			jsonPayload := make(map[string]interface{})
			err = json.Unmarshal([]byte(logText), &jsonPayload)
			if err == nil {
				log.JSONPayload = jsonPayload
			} else {
				log.TextPayload = logText
			}

			podLogs = append(podLogs, log)
		}

		if request.HeadLines != nil {
			endIndex := *request.HeadLines
			// Check against slice index out of bounds
			if *request.HeadLines > int64(len(podLogs)) {
				endIndex = int64(len(podLogs))
			} else if *request.HeadLines < 0 {
				endIndex = 0
			}
			allPodLogs = append(allPodLogs, podLogs[:endIndex]...)
		} else if request.TailLines != nil {
			startIndex := int64(len(podLogs)) - *request.TailLines
			// Check against slice index out of bounds
			if startIndex < 0 {
				startIndex = 0
			}
			allPodLogs = append(allPodLogs, podLogs[startIndex:]...)
		} else {
			allPodLogs = append(allPodLogs, podLogs...)
		}
	}

	// Sort all the logs by timestamp in ascending order
	sort.Slice(allPodLogs, func(i, j int) bool {
		return allPodLogs[i].Timestamp.Before(allPodLogs[j].Timestamp)
	})

	return allPodLogs, nil
}
