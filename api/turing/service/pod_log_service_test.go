package service

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	mock "github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/cluster/mocks"
	"github.com/caraml-dev/turing/api/turing/cluster/servicebuilder"
	"github.com/caraml-dev/turing/api/turing/models"
)

func TestConvertPodLogsToV2(t *testing.T) {
	namespace := "namespace"
	environment := "environment"
	loggingURL := "wwww.example.com"

	tests := map[string]struct {
		legacyPodLogs []*PodLog
		want          *PodLogsV2
	}{
		"success | nominal": {
			legacyPodLogs: []*PodLog{
				{
					Timestamp:     time.Date(2020, 7, 7, 7, 0, 5, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "No this is patrick",
				},
			},
			want: &PodLogsV2{
				Environment: environment,
				Namespace:   namespace,
				LoggingURL:  loggingURL,
				Logs: []*PodLogV2{
					{
						Timestamp:   time.Date(2020, 7, 7, 7, 0, 5, 0, time.UTC),
						PodName:     "json-payload",
						TextPayload: "No this is patrick",
					},
				},
			},
		},
		"success | no log entry": {
			legacyPodLogs: []*PodLog{},
			want: &PodLogsV2{
				Environment: environment,
				Namespace:   namespace,
				LoggingURL:  loggingURL,
				Logs:        []*PodLogV2{},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := ConvertPodLogsToV2(namespace, environment, loggingURL, tt.legacyPodLogs)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("ConvertPodLogsToV2() got = %v, want %v", got, tt.want)
				t.Log(cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestPodLogServiceListPodLogs(t *testing.T) {
	sinceTime := time.Date(2020, 7, 7, 7, 0, 0, 0, time.UTC)
	sinceTimeMinus1Sec := time.Date(2020, 7, 7, 6, 59, 59, 0, time.UTC)
	sinceTimeV1Minus1Sec := metav1.Time{Time: sinceTimeMinus1Sec}
	headLines := int64(2)
	tailLines := int64(1)

	controller := &mocks.Controller{}
	controller.
		On("ListPods", mock.Anything, "namespace", "serving.knative.dev/service=router1-turing-router-1").
		Return(&corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "json-payload"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "user-container"},
							{Name: "queue-proxy"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "text-payload"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "user-container"},
							{Name: "queue-proxy"},
						},
					},
				},
			},
		}, nil)
	controller.
		On("ListPods", mock.Anything, "listpods-error", "serving.knative.dev/service=router1-turing-router-1").
		Return(nil, errors.New(""))
	controller.
		On("ListPods", mock.Anything, "listpodlogs-error", "serving.knative.dev/service=router1-turing-router-1").
		Return(&corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "listpodlogs-error"},
					Spec:       corev1.PodSpec{Containers: []corev1.Container{}},
				},
			},
		}, nil)
	controller.
		On("ListPodLogs", mock.Anything, "namespace", "json-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, SinceTime: &sinceTimeV1Minus1Sec}).
		Return(io.NopCloser(bytes.NewBufferString(`2020-07-07T06:59:59Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:05Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:10Z {"foo":"bar", "baz": 10}`)), nil)
	controller.
		On("ListPodLogs", mock.Anything, "namespace", "json-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true}).
		Return(io.NopCloser(bytes.NewBufferString(`2020-07-07T06:59:59Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:05Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:10Z {"foo":"bar", "baz": 10}`)), nil)
	controller.
		On("ListPodLogs", mock.Anything, "namespace", "json-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, TailLines: &tailLines}).
		Return(io.NopCloser(bytes.NewBufferString(`2020-07-07T07:00:05Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:10Z {"foo":"bar", "baz": 10}`)), nil)
	controller.
		On("ListPodLogs", mock.Anything, "namespace", "text-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, SinceTime: &sinceTimeV1Minus1Sec}).
		Return(io.NopCloser(bytes.NewBufferString(`2020-07-07T08:00:05Z line1
2020-07-07T08:00:10Z line2
invalidtimestamp line3

2020-07-07T08:00:00Z `)), nil)
	controller.
		On("ListPodLogs", mock.Anything, "namespace", "text-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, TailLines: &tailLines}).
		Return(io.NopCloser(bytes.NewBufferString(`2020-07-07T08:00:05Z line1
2020-07-07T08:00:10Z line2
invalidtimestamp line3

2020-07-07T08:00:00Z `)), nil)
	controller.
		On("ListPodLogs", mock.Anything, "namespace", "text-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true}).
		Return(io.NopCloser(bytes.NewBufferString(`2020-07-07T08:00:05Z line1
2020-07-07T08:00:10Z line2
invalidtimestamp line3

2020-07-07T08:00:00Z `)), nil)
	controller.
		On("ListPodLogs", mock.Anything, "listpodlogs-error", "listpodlogs-error",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true}).
		Return(nil, errors.New(""))
	clusterControllers := map[string]cluster.Controller{"environment": controller}

	routerVersion := &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1}

	tests := []struct {
		name    string
		args    PodLogRequest
		want    []*PodLog
		wantErr bool
	}{
		{
			name: "expected arguments with headlines and taillines",
			args: PodLogRequest{
				Namespace:        "namespace",
				DefaultContainer: cluster.KnativeUserContainerName,
				Environment:      "environment",
				LabelSelectors: []LabelSelector{
					{
						Key:   cluster.KnativeServiceLabelKey,
						Value: servicebuilder.GetComponentName(routerVersion, "router"),
					},
				},
				SinceTime: &sinceTime,
				HeadLines: &headLines,
				TailLines: &tailLines,
			},
			want: []*PodLog{
				{
					Timestamp:     time.Date(2020, 7, 7, 7, 0, 5, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "",
					JSONPayload:   map[string]interface{}{"foo": "bar", "baz": float64(5)},
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 7, 0, 10, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "",
					JSONPayload:   map[string]interface{}{"foo": "bar", "baz": float64(10)},
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 8, 0, 5, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "text-payload",
					ContainerName: "user-container",
					TextPayload:   "line1",
					JSONPayload:   nil,
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 8, 0, 10, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "text-payload",
					ContainerName: "user-container",
					TextPayload:   "line2",
					JSONPayload:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "expected arguments with taillines only",
			args: PodLogRequest{
				Namespace:        "namespace",
				DefaultContainer: cluster.KnativeUserContainerName,
				Environment:      "environment",
				LabelSelectors: []LabelSelector{
					{
						Key:   cluster.KnativeServiceLabelKey,
						Value: servicebuilder.GetComponentName(routerVersion, "router"),
					},
				},
				TailLines: &tailLines,
			},
			want: []*PodLog{
				{
					Timestamp:     time.Date(2020, 7, 7, 7, 0, 10, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "",
					JSONPayload:   map[string]interface{}{"foo": "bar", "baz": float64(10)},
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 8, 0, 10, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "text-payload",
					ContainerName: "user-container",
					TextPayload:   "line2",
					JSONPayload:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "expected arguments no headlines and taillines",
			args: PodLogRequest{
				Namespace:        "namespace",
				DefaultContainer: cluster.KnativeUserContainerName,
				Environment:      "environment",
				LabelSelectors: []LabelSelector{
					{
						Key:   cluster.KnativeServiceLabelKey,
						Value: servicebuilder.GetComponentName(routerVersion, "router"),
					},
				},
			},
			want: []*PodLog{
				{
					Timestamp:     time.Date(2020, 7, 7, 6, 59, 59, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "",
					JSONPayload:   map[string]interface{}{"foo": "bar", "baz": float64(5)},
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 7, 0, 5, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "",
					JSONPayload:   map[string]interface{}{"foo": "bar", "baz": float64(5)},
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 7, 0, 10, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "json-payload",
					ContainerName: "user-container",
					TextPayload:   "",
					JSONPayload:   map[string]interface{}{"foo": "bar", "baz": float64(10)},
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 8, 0, 5, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "text-payload",
					ContainerName: "user-container",
					TextPayload:   "line1",
					JSONPayload:   nil,
				},
				{
					Timestamp:     time.Date(2020, 7, 7, 8, 0, 10, 0, time.UTC),
					Environment:   "environment",
					Namespace:     "namespace",
					PodName:       "text-payload",
					ContainerName: "user-container",
					TextPayload:   "line2",
					JSONPayload:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "controller listpods error",
			args: PodLogRequest{
				Namespace:        "listpods-error",
				DefaultContainer: cluster.KnativeUserContainerName,
				Environment:      "environment",
				LabelSelectors: []LabelSelector{
					{
						Key:   cluster.KnativeServiceLabelKey,
						Value: servicebuilder.GetComponentName(routerVersion, "router"),
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "controller for environment not found",
			args: PodLogRequest{
				Namespace:        "namespace",
				DefaultContainer: cluster.KnativeUserContainerName,
				Environment:      "environment-not-found",
				LabelSelectors: []LabelSelector{
					{
						Key:   cluster.KnativeServiceLabelKey,
						Value: servicebuilder.GetComponentName(routerVersion, "router"),
					},
				},
				SinceTime: &sinceTime,
				HeadLines: &headLines,
				TailLines: &tailLines,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podLogService{clusterControllers: clusterControllers}
			got, err := s.ListPodLogs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPodLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("ListPodLogs() got = %v, want %v", got, tt.want)
				t.Log(cmp.Diff(got, tt.want))
			}
		})
	}
}
