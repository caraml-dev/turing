package service

import (
	"reflect"
	"testing"

	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	mockImgBuilder "github.com/caraml-dev/turing/api/turing/imagebuilder/mocks"
	"github.com/caraml-dev/turing/api/turing/models"
	mock "github.com/stretchr/testify/mock"
)

func Test_ensemblerImagesService_ListImages(t *testing.T) {
	type args struct {
		project    *mlp.Project
		ensembler  *models.PyFuncEnsembler
		runnerType models.EnsemblerRunnerType
	}
	tests := []struct {
		name                         string
		ensemblerJobImageBuilder     func() *mockImgBuilder.ImageBuilder
		ensemblerServiceImageBuilder func() *mockImgBuilder.ImageBuilder
		args                         args
		want                         []imagebuilder.EnsemblerImage
		wantErr                      bool
	}{
		{
			name: "success - ensembler job image",
			ensemblerJobImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				ib.On("GetEnsemblerImage", mock.Anything, mock.Anything).
					Return(imagebuilder.EnsemblerImage{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
						Exists:              true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					}, nil)
				return ib
			},
			ensemblerServiceImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				return ib
			},
			args: args{
				project: &mlp.Project{
					ID:   1,
					Name: "myproject",
				},
				ensembler: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				},
				runnerType: models.EnsemblerRunnerTypeJob,
			},
			want: []imagebuilder.EnsemblerImage{
				{
					ProjectID:           models.ID(1),
					EnsemblerID:         models.ID(1),
					EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
					ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
					Exists:              true,
					JobStatus: imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success - ensembler service image",
			ensemblerJobImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				return ib
			},
			ensemblerServiceImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				ib.On("GetEnsemblerImage", mock.Anything, mock.Anything).
					Return(imagebuilder.EnsemblerImage{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
						Exists:              true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					}, nil)
				return ib
			},
			args: args{
				project: &mlp.Project{
					ID:   1,
					Name: "myproject",
				},
				ensembler: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				},
				runnerType: models.EnsemblerRunnerTypeService,
			},
			want: []imagebuilder.EnsemblerImage{
				{
					ProjectID:           models.ID(1),
					EnsemblerID:         models.ID(1),
					EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
					ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
					Exists:              true,
					JobStatus: imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success - both ensembler job and service image",
			ensemblerJobImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				ib.On("GetEnsemblerImage", mock.Anything, mock.Anything).
					Return(imagebuilder.EnsemblerImage{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
						Exists:              true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					}, nil)
				return ib
			},
			ensemblerServiceImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				ib.On("GetEnsemblerImage", mock.Anything, mock.Anything).
					Return(imagebuilder.EnsemblerImage{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
						Exists:              true,
					}, nil)
				ib.On("GetImageBuildingJobStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					}, nil)
				return ib
			},
			args: args{
				project: &mlp.Project{
					ID:   1,
					Name: "myproject",
				},
				ensembler: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				},
			},
			want: []imagebuilder.EnsemblerImage{
				{
					ProjectID:           models.ID(1),
					EnsemblerID:         models.ID(1),
					EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
					ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
					Exists:              true,
					JobStatus: imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					},
				},
				{
					ProjectID:           models.ID(1),
					EnsemblerID:         models.ID(1),
					EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
					ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
					Exists:              true,
					JobStatus: imagebuilder.JobStatus{
						State: imagebuilder.JobStateSucceeded,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ensemblerImagesService{
				ensemblerJobImageBuilder:     tt.ensemblerJobImageBuilder(),
				ensemblerServiceImageBuilder: tt.ensemblerServiceImageBuilder(),
			}
			got, err := s.ListImages(tt.args.project, tt.args.ensembler, tt.args.runnerType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensemblerImagesService.ListImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ensemblerImagesService.ListImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ensemblerImagesService_BuildImage(t *testing.T) {
	type args struct {
		project    *mlp.Project
		ensembler  *models.PyFuncEnsembler
		runnerType models.EnsemblerRunnerType
	}
	tests := []struct {
		name                         string
		ensemblerJobImageBuilder     func() *mockImgBuilder.ImageBuilder
		ensemblerServiceImageBuilder func() *mockImgBuilder.ImageBuilder
		args                         args
		wantErr                      bool
	}{
		{
			name: "success - build ensembler job image",
			ensemblerJobImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				ib.On("BuildImage", mock.Anything).
					Return("ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123", nil)
				return ib
			},
			ensemblerServiceImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				return ib
			},
			args: args{
				project: &mlp.Project{
					ID:   1,
					Name: "myproject",
				},
				ensembler: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				},
				runnerType: models.EnsemblerRunnerTypeJob,
			},
			wantErr: false,
		},
		{
			name: "success - build ensembler service image",
			ensemblerJobImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				return ib
			},
			ensemblerServiceImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				ib.On("BuildImage", mock.Anything).
					Return("ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123", nil)
				return ib
			},
			args: args{
				project: &mlp.Project{
					ID:   1,
					Name: "myproject",
				},
				ensembler: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				},
				runnerType: models.EnsemblerRunnerTypeService,
			},
			wantErr: false,
		},
		{
			name: "invalid runner type",
			ensemblerJobImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				return ib
			},
			ensemblerServiceImageBuilder: func() *mockImgBuilder.ImageBuilder {
				ib := &mockImgBuilder.ImageBuilder{}
				return ib
			},
			args: args{
				project: &mlp.Project{
					ID:   1,
					Name: "myproject",
				},
				ensembler: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ensemblerImagesService{
				ensemblerJobImageBuilder:     tt.ensemblerJobImageBuilder(),
				ensemblerServiceImageBuilder: tt.ensemblerServiceImageBuilder(),
			}
			if err := s.BuildImage(tt.args.project, tt.args.ensembler, tt.args.runnerType); (err != nil) != tt.wantErr {
				t.Errorf("ensemblerImagesService.BuildImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
