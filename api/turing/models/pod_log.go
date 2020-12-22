package models

import "time"

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

