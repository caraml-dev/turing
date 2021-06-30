package imagebuilder

import "github.com/pkg/errors"

var (
	// ErrUnableToGetImageRef is an error that failed to check if a previous image exists.
	ErrUnableToGetImageRef = errors.New("error checking previous image")

	// ErrUnableToGetJobStatus is an error that failed to get job status.
	ErrUnableToGetJobStatus = errors.New("unknown kaniko builder status")

	// ErrUnableToBuildImage is an error that failed to build an OCI image.
	ErrUnableToBuildImage = errors.New("error building OCI image")

	// ErrDeleteFailedJob is an error that the OCI image building process has failed.
	ErrDeleteFailedJob = errors.New("error deleting kaniko builder")

	// ErrUnableToParseKanikoResource is an error that failed to parse kubernetes resources for kaniko builder.
	ErrUnableToParseKanikoResource = errors.New("error parsing kaniko resources")

	// ErrTimeoutBuildingImage is an error that the image build timed out.
	ErrTimeoutBuildingImage = errors.New("timeout building pyfunc image")
)
