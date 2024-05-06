import json
import logging

import pytest
from urllib3_mock import Responses

import tests
import turing.ensembler
import turing.generated
import turing.generated.models
from turing.ensembler_image import EnsemblerImage

responses = Responses("requests.packages.urllib3")


@pytest.fixture(scope="module", name="responses")
def _responses():
    return responses


@responses.activate
@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
def test_list_images(turing_api, use_google_oauth, active_project, pyfunc_ensembler):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    ensembler_id = 1

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers/{ensembler_id}",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    ensembler = turing.PyFuncEnsembler.get_by_id(ensembler_id)

    job_image = turing.generated.models.EnsemblerImage(
        project_id=active_project.id,
        ensembler_id=ensembler_id,
        runner_type=turing.generated.models.EnsemblerImageRunnerType("job"),
        image_ref=f"ghcr.io/caraml-dev/turing/ensembler-jobs/{active_project.name}-ensembler-1:latest",
        exists=True,
        image_building_job_status=turing.generated.models.ImageBuildingJobStatus(
            state=turing.generated.models.ImageBuildingJobState("succeeded"),
        ),
    )

    service_image = turing.generated.models.EnsemblerImage(
        project_id=active_project.id,
        ensembler_id=ensembler_id,
        runner_type=turing.generated.models.EnsemblerImageRunnerType("service"),
        image_ref=f"ghcr.io/caraml-dev/turing/ensembler-services/{active_project.name}-ensembler-1:latest",
        exists=True,
        image_building_job_status=turing.generated.models.ImageBuildingJobStatus(
            state=turing.generated.models.ImageBuildingJobState("succeeded"),
        ),
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers/{ensembler_id}/images?runner_type=job",
        match_querystring=True,
        body=json.dumps([job_image], default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    images = turing.EnsemblerImage.list(ensembler=ensembler, runner_type="job")
    assert len(images) == 1
    assert EnsemblerImage.from_open_api(job_image) in images
    assert EnsemblerImage.from_open_api(service_image) not in images

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers/{ensembler_id}/images",
        body=json.dumps([job_image, service_image], default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    images = turing.EnsemblerImage.list(ensembler=ensembler)
    assert len(images) == 2
    assert EnsemblerImage.from_open_api(job_image) in images
    assert EnsemblerImage.from_open_api(service_image) in images


@responses.activate
@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
def test_create_image(turing_api, use_google_oauth, active_project, pyfunc_ensembler):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    ensembler_id = 1

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers/{ensembler_id}",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    ensembler = turing.PyFuncEnsembler.get_by_id(ensembler_id)

    responses.add(
        method="PUT",
        url=f"/v1/projects/{active_project.id}/ensemblers/{ensembler_id}/images",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=201,
        content_type="application/json",
    )

    job_image = turing.generated.models.EnsemblerImage(
        project_id=active_project.id,
        ensembler_id=ensembler_id,
        runner_type=turing.generated.models.EnsemblerImageRunnerType("job"),
        image_ref=f"ghcr.io/caraml-dev/turing/ensembler-jobs/{active_project.name}-ensembler-1:latest",
        exists=True,
        image_building_job_status=turing.generated.models.ImageBuildingJobStatus(
            state=turing.generated.models.ImageBuildingJobState("succeeded"),
        ),
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers/{ensembler_id}/images?runner_type=job",
        match_querystring=True,
        body=json.dumps([job_image], default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    turing.EnsemblerImage.create(
        ensembler=ensembler,
        runner_type=turing.generated.models.EnsemblerImageRunnerType("job"),
    )

    images = turing.EnsemblerImage.list(ensembler=ensembler, runner_type="job")
    assert len(images) == 1
    assert EnsemblerImage.from_open_api(job_image) in images
