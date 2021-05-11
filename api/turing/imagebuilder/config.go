package imagebuilder

import (
	"time"
)

// ImageConfig provides the configuration used for the OCI image building.
type ImageConfig struct {
	Registry             string
	BaseImageRef         string
	BuildNamespace       string
	BuildContextURI      string
	DockerfileFilePath   string
	BuildTimeoutDuration time.Duration
}

// ResourceRequestsLimits contains the Kubernetes resource request and limits for kaniko
type ResourceRequestsLimits struct {
	Requests Resource
	Limits   Resource
}

// Resource contains the Kubernetes resource request and limits
type Resource struct {
	CPU    string
	Memory string
}

// KanikoConfig provides the configuration used for the Kaniko image.
type KanikoConfig struct {
	Image                  string
	ImageVersion           string
	ResourceRequestsLimits ResourceRequestsLimits
}
