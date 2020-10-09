package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/cluster/servicebuilder"
	logger "github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodLogService interface {
	// ListPodLogs retrieves logs from user-container (default) container from all pods running the router.
	ListPodLogs(
		project *mlp.Project,
		router *models.Router,
		routerVersion *models.RouterVersion,
		componentType string,
		opts *PodLogOptions,
	) ([]*PodLog, error)
}

// PodLog represents a single log line from a container running in Kubernetes.
// If the log message is in JSON format, JSONPayload must be populated with the
// structured JSON log message. Else, TextPayload must be populated with log text.
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
	ContainerName string `json:"container_name"`
	// Log in text format, either TextPayload or JSONPayload will be set but not both
	TextPayload string `json:"text_payload,omitempty"`
	// Log in JSON format, either TextPayload or JSONPayload will be set but not both
	JSONPayload map[string]interface{} `json:"json_payload,omitempty"`
}

type PodLogOptions struct {
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

type podLogService struct {
	clusterControllers map[string]cluster.Controller
}

func NewPodLogService(clusterControllers map[string]cluster.Controller) PodLogService {
	return &podLogService{clusterControllers: clusterControllers}
}

func (s *podLogService) ListPodLogs(
	project *mlp.Project,
	router *models.Router,
	routerVersion *models.RouterVersion,
	componentType string,
	opts *PodLogOptions,
) ([]*PodLog, error) {
	// If both TailLines and Headlines are set, TailLines is ignored
	if opts.TailLines != nil && opts.HeadLines != nil {
		opts.TailLines = nil
	}

	controller, ok := s.clusterControllers[router.EnvironmentName]
	if !ok {
		return nil, fmt.Errorf("cluster controller for environment '%s' not found", router.EnvironmentName)
	}

	namespace := servicebuilder.GetNamespace(project)
	labelSelector := cluster.KnativeServiceLabelKey + "=" +
		servicebuilder.GetComponentName(routerVersion, componentType)
	pods, err := controller.ListPods(namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	allPodLogs := make([]*PodLog, 0)
	for _, p := range pods.Items {
		logOpts := &v1.PodLogOptions{
			Container:  opts.Container,
			Previous:   opts.Previous,
			Timestamps: true,
		}

		// Only send tailLines argument to Kube API server if sinceTime is not set because Kube API will prioritize
		// tailLines over sinceTime and we want these options to be AND-ed together.
		// The log items returned by Kube API will later be trimmed separately according to the tailLines argument.
		if opts.SinceTime == nil && opts.TailLines != nil {
			logOpts.TailLines = opts.TailLines
		}

		// Set default container if not set
		if logOpts.Container == "" {
			logOpts.Container = cluster.KnativeUserContainerName
		}

		if opts.SinceTime != nil {
			// Note that the requested sinceTime sent to Kubernetes API server is subtracted by 1 second
			// because Kubernetes API server only accepts sinceTime granularity to the second so we will
			// miss some logs within the second without this subtraction.
			logOpts.SinceTime = &metav1.Time{Time: opts.SinceTime.Add(-time.Second)}
		}

		stream, err := controller.ListPodLogs(namespace, p.Name, logOpts)
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
			if opts.SinceTime != nil && (timestamp == *opts.SinceTime || timestamp.Before(*opts.SinceTime)) {
				continue
			}

			log := &PodLog{
				Timestamp:     timestamp,
				Environment:   router.EnvironmentName,
				Namespace:     namespace,
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

		if opts.HeadLines != nil {
			endIndex := *opts.HeadLines
			// Check against slice index out of bounds
			if *opts.HeadLines > int64(len(podLogs)) {
				endIndex = int64(len(podLogs))
			} else if *opts.HeadLines < 0 {
				endIndex = 0
			}
			allPodLogs = append(allPodLogs, podLogs[:endIndex]...)
		} else if opts.TailLines != nil {
			startIndex := int64(len(podLogs)) - *opts.TailLines
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
