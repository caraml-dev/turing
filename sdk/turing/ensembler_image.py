import logging
from typing import List

import turing
import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec


@ApiObjectSpec(turing.generated.models.EnsemblerImage)
class EnsemblerImage(ApiObject):
    """
    API entity for Ensembler Image
    """

    def __init__(
        self,
        project_id: int,
        ensembler_id: int,
        runner_type: turing.generated.models.EnsemblerImageRunnerType,
        image_ref: str,
        exists: bool,
        image_building_job_status: turing.generated.models.ImageBuildingJobStatus,
        **kwargs,
    ):
        super(EnsemblerImage, self).__init__(**kwargs)
        self._project_id = project_id
        self._ensembler_id = ensembler_id
        self._runner_type = runner_type
        self._image_ref = image_ref
        self._exists = exists
        self._image_building_job_status = image_building_job_status

    @property
    def project_id(self) -> int:
        return self._project_id

    @property
    def ensembler_id(self) -> int:
        return self._ensembler_id

    @property
    def runner_type(self) -> turing.generated.models.EnsemblerImageRunnerType:
        return self._runner_type

    @property
    def image_ref(self) -> str:
        return self._image_ref

    @property
    def exists(self) -> bool:
        return self._exists

    @property
    def image_building_job_status(
        self,
    ) -> turing.generated.models.ImageBuildingJobStatus:
        return self._image_building_job_status

    def list(
        ensembler: turing.generated.models.Ensembler,
        runner_type: turing.generated.models.EnsemblerImageRunnerType = None,
    ) -> List["EnsemblerImage"]:
        """
        List all Docker images for the ensembler

        :param ensembler: Ensembler object
        :param runner_type: (optional) Runner type of image building job used to filter the images. (default: None, options: [None, 'job', 'service'])

        :return: List of EnsemblerImage objects
        """
        response = turing.active_session.list_ensembler_images(
            ensembler=ensembler, runner_type=runner_type
        )
        return [EnsemblerImage.from_open_api(item) for item in response.value]

    def create(
        ensembler: turing.generated.models.Ensembler,
        runner_type: turing.generated.models.EnsemblerImageRunnerType,
    ) -> "EnsemblerImage":
        """
        Create a new Docker image for the ensembler

        :param ensembler: Ensembler object
        :param runner_type: Runner type of image building job (options: ['job', 'service'])
        """
        image = turing.active_session.create_ensembler_image(
            ensembler=ensembler, runner_type=runner_type
        )
        return EnsemblerImage.from_open_api(image)
