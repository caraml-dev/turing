package service

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/cluster/mocks"
	"github.com/gojek/turing/api/turing/models"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodLogServiceListPodLogs(t *testing.T) {
	sinceTime := time.Date(2020, 7, 7, 7, 0, 0, 0, time.UTC)
	sinceTimeMinus1Sec := time.Date(2020, 7, 7, 6, 59, 59, 0, time.UTC)
	sinceTimeV1Minus1Sec := metav1.Time{Time: sinceTimeMinus1Sec}
	headLines := int64(2)
	tailLines := int64(1)

	controller := &mocks.Controller{}
	controller.
		On("ListPods", "namespace", "serving.knative.dev/service=router1-turing-router-1").
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
		On("ListPods", "listpods-error", "serving.knative.dev/service=router1-turing-router-1").
		Return(nil, errors.New(""))
	controller.
		On("ListPods", "listpodlogs-error", "serving.knative.dev/service=router1-turing-router-1").
		Return(&corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "listpodlogs-error"},
					Spec:       corev1.PodSpec{Containers: []corev1.Container{}},
				},
			},
		}, nil)
	controller.
		On("ListPodLogs", "namespace", "json-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, SinceTime: &sinceTimeV1Minus1Sec}).
		Return(ioutil.NopCloser(bytes.NewBufferString(`2020-07-07T06:59:59Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:05Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:10Z {"foo":"bar", "baz": 10}`)), nil)
	controller.
		On("ListPodLogs", "namespace", "json-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true}).
		Return(ioutil.NopCloser(bytes.NewBufferString(`2020-07-07T06:59:59Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:05Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:10Z {"foo":"bar", "baz": 10}`)), nil)
	controller.
		On("ListPodLogs", "namespace", "json-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, TailLines: &tailLines}).
		Return(ioutil.NopCloser(bytes.NewBufferString(`2020-07-07T07:00:05Z {"foo":"bar", "baz": 5}
2020-07-07T07:00:10Z {"foo":"bar", "baz": 10}`)), nil)
	controller.
		On("ListPodLogs", "namespace", "text-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, SinceTime: &sinceTimeV1Minus1Sec}).
		Return(ioutil.NopCloser(bytes.NewBufferString(`2020-07-07T08:00:05Z line1
2020-07-07T08:00:10Z line2
invalidtimestamp line3

2020-07-07T08:00:00Z `)), nil)
	controller.
		On("ListPodLogs", "namespace", "text-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true, TailLines: &tailLines}).
		Return(ioutil.NopCloser(bytes.NewBufferString(`2020-07-07T08:00:05Z line1
2020-07-07T08:00:10Z line2
invalidtimestamp line3

2020-07-07T08:00:00Z `)), nil)
	controller.
		On("ListPodLogs", "namespace", "text-payload",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true}).
		Return(ioutil.NopCloser(bytes.NewBufferString(`2020-07-07T08:00:05Z line1
2020-07-07T08:00:10Z line2
invalidtimestamp line3

2020-07-07T08:00:00Z `)), nil)
	controller.
		On("ListPodLogs", "listpodlogs-error", "listpodlogs-error",
			&corev1.PodLogOptions{Container: "user-container", Timestamps: true}).
		Return(nil, errors.New(""))
	clusterControllers := map[string]cluster.Controller{"environment": controller}

	type args struct {
		project       *client.Project
		router        *models.Router
		routerVersion *models.RouterVersion
		componentType string
		opts          *PodLogOptions
	}
	tests := []struct {
		name    string
		args    args
		want    []*PodLog
		wantErr bool
	}{
		{
			name: "expected arguments with headlines and taillines",
			args: args{
				project:       &client.Project{Name: "namespace"},
				router:        &models.Router{Name: "router1", EnvironmentName: "environment"},
				routerVersion: &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1},
				componentType: "router",
				opts:          &PodLogOptions{SinceTime: &sinceTime, HeadLines: &headLines, TailLines: &tailLines},
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
			args: args{
				project:       &client.Project{Name: "namespace"},
				router:        &models.Router{Name: "router1", EnvironmentName: "environment"},
				routerVersion: &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1},
				componentType: "router",
				opts:          &PodLogOptions{TailLines: &tailLines},
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
			args: args{
				project:       &client.Project{Name: "namespace"},
				router:        &models.Router{Name: "router1", EnvironmentName: "environment"},
				routerVersion: &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1},
				componentType: "router",
				opts:          &PodLogOptions{},
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
			args: args{
				project:       &client.Project{Name: "listpods-error"},
				router:        &models.Router{Name: "router1", EnvironmentName: "environment"},
				routerVersion: &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1},
				componentType: "router",
				opts:          &PodLogOptions{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "controller listpodlogs error",
			args: args{
				project:       &client.Project{Name: "listpodlogs-error"},
				router:        &models.Router{Name: "router1", EnvironmentName: "environment"},
				routerVersion: &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1},
				componentType: "router",
				opts:          &PodLogOptions{},
			},
			want:    []*PodLog{},
			wantErr: false,
		},
		{
			name: "controller for environment not found",
			args: args{
				project:       &client.Project{Name: "namespace"},
				router:        &models.Router{Name: "router1", EnvironmentName: "environment-not-found"},
				routerVersion: &models.RouterVersion{Router: &models.Router{Name: "router1"}, Version: 1},
				componentType: "router",
				opts:          &PodLogOptions{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &podLogService{clusterControllers: clusterControllers}
			got, err := s.ListPodLogs(tt.args.project, tt.args.router, tt.args.routerVersion,
				tt.args.componentType, tt.args.opts)
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
