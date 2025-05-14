import json
import logging
from unittest.mock import MagicMock, patch

import pytest
from urllib3_mock import Responses

import tests
import turing.generated
import turing.generated.models
from turing.ensembler_image import EnsemblerImage

@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
def test_list_images(turing_api, use_google_oauth, project, pyfunc_ensembler, active_project_magic_mock):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)
    
        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        ensembler_id = 1

        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers/{ensembler_id}"
        mock_response.data = json.dumps(pyfunc_ensembler, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'

        mock_request.return_value = mock_response

        ensembler = turing.PyFuncEnsembler.get_by_id(ensembler_id)

        job_image = turing.generated.models.EnsemblerImage(
            project_id=project.id,
            ensembler_id=ensembler_id,
            runner_type=turing.generated.models.EnsemblerImageRunnerType("job"),
            image_ref=f"ghcr.io/caraml-dev/turing/ensembler-jobs/{project.name}-ensembler-1:latest",
            exists=True,
            image_building_job_status=turing.generated.models.ImageBuildingJobStatus(
                state=turing.generated.models.ImageBuildingJobState("succeeded"),
            ),
        )

        service_image = turing.generated.models.EnsemblerImage(
            project_id=project.id,
            ensembler_id=ensembler_id,
            runner_type=turing.generated.models.EnsemblerImageRunnerType("service"),
            image_ref=f"ghcr.io/caraml-dev/turing/ensembler-services/{project.name}-ensembler-1:latest",
            exists=True,
            image_building_job_status=turing.generated.models.ImageBuildingJobStatus(
                state=turing.generated.models.ImageBuildingJobState("succeeded"),
            ),
        )
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers/{ensembler_id}/images?runner_type=job"
        mock_response.data = json.dumps([job_image], default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'

        mock_request.return_value = mock_response

        images = turing.EnsemblerImage.list(ensembler=ensembler, runner_type="job")
        assert len(images) == 1
        assert EnsemblerImage.from_open_api(job_image) in images
        assert EnsemblerImage.from_open_api(service_image) not in images

        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers/{ensembler_id}/images"
        mock_response.data = json.dumps([job_image, service_image], default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'

        mock_request.return_value = mock_response

        images = turing.EnsemblerImage.list(ensembler=ensembler)
        assert len(images) == 2
        assert EnsemblerImage.from_open_api(job_image) in images
        assert EnsemblerImage.from_open_api(service_image) in images

@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
def test_create_image(turing_api, use_google_oauth, project, pyfunc_ensembler, active_project_magic_mock):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)

        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        ensembler_id = 1

        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers/{ensembler_id}"
        mock_response.data = json.dumps(pyfunc_ensembler, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        ensembler = turing.PyFuncEnsembler.get_by_id(ensembler_id)

        job_image = turing.generated.models.EnsemblerImage(
            project_id=project.id,
            ensembler_id=ensembler_id,
            runner_type=turing.generated.models.EnsemblerImageRunnerType("job"),
            image_ref=f"ghcr.io/caraml-dev/turing/ensembler-jobs/{project.name}-ensembler-1:latest",
            exists=True,
            image_building_job_status=turing.generated.models.ImageBuildingJobStatus(
                state=turing.generated.models.ImageBuildingJobState("succeeded"),
            ),
        )
         
        mock_response_1 = MagicMock()
        mock_response_1.method = "PUT"
        mock_response_1.status = 201
        mock_response_1.path = f"/v1/projects/{project.id}/ensemblers/{ensembler_id}/images"
        mock_response_1.data = json.dumps(pyfunc_ensembler, default=tests.json_serializer).encode('utf-8')
        mock_response_1.getheader.return_value = 'application/json'
        
        mock_response_2 = MagicMock()
        mock_response_2.method = "GET"
        mock_response_2.status = 200
        mock_response_2.path = f"/v1/projects/{project.id}/ensemblers/{ensembler_id}/images?runner_type=job"
        mock_response_2.data = json.dumps([job_image], default=tests.json_serializer).encode('utf-8')
        mock_response_2.getheader.return_value = 'application/json'
        
        mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_2]

        turing.EnsemblerImage.create(
            ensembler=ensembler,
            runner_type=turing.generated.models.EnsemblerImageRunnerType("job"),
        )

        mock_request.return_value = mock_response_2

        images = turing.EnsemblerImage.list(ensembler=ensembler, runner_type="job")
        assert len(images) == 1
        assert EnsemblerImage.from_open_api(job_image) in images