package models

// Enricher contains the configuration for a preprocessor for
// a Turing Router.
type Enricher struct {
	Model
	// Fully qualified docker image string used by the enricher, in the
	// format registry/repository:version.
	Image string `json:"image"`
	// Resource requests  for the deployment of the enricher.
	ResourceRequest *ResourceRequest `json:"resource_request"`
	// Autoscaling policy for the enricher
	AutoscalingPolicy AutoscalingPolicy `json:"autoscaling_policy" validate:"omitempty,dive"`
	// Endpoint to query.
	Endpoint string `json:"endpoint"`
	// Request timeout as a valid quantity string.
	Timeout string `json:"timeout"`
	// Port to query.
	Port int `json:"port"`
	// Environment variables to inject into the pod.
	Env EnvVars `json:"env"`
	// (optional) ServiceAccount specifies the name of the secret registered in the MLP project containing the service
	// account. The service account will be mounted into the container and the env variable
	// GOOGLE_APPLICATION_CREDENTIALS will point to the service account file.
	ServiceAccount string `json:"service_account"`
}
