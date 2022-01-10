import re
import turing.generated.models
from turing.generated.model_utils import OpenApiModel


class ResourceRequest:
    min_allowed_replica = 0
    max_allowed_replica = 20

    def __init__(self, min_replica: int, max_replica: int, cpu_request: str, memory_request: str):
        """
        Method to create a new ResourceRequest object

        :param min_replica: min number of replicas available
        :param max_replica: max number of replicas available
        :param cpu_request: total amount of CPU available
        :param memory_request: total amount of RAM available
        """
        ResourceRequest._verify_min_max_replica(min_replica, max_replica)
        ResourceRequest._verify_cpu_request(cpu_request)
        ResourceRequest._verify_memory_request(memory_request)

        self._min_replica = min_replica
        self._max_replica = max_replica
        self._cpu_request = cpu_request
        self._memory_request = memory_request

    @property
    def min_replica(self) -> int:
        return self._min_replica

    @min_replica.setter
    def min_replica(self, min_replica):
        ResourceRequest._verify_min_max_replica(min_replica, self._max_replica)
        self._min_replica = min_replica

    @property
    def max_replica(self) -> int:
        return self._max_replica

    @max_replica.setter
    def max_replica(self, max_replica):
        ResourceRequest._verify_min_max_replica(self._min_replica, max_replica)
        self._max_replica = max_replica

    @property
    def cpu_request(self) -> str:
        return self._cpu_request

    @cpu_request.setter
    def cpu_request(self, cpu_request):
        ResourceRequest._verify_cpu_request(cpu_request)
        self._cpu_request = cpu_request

    @property
    def memory_request(self) -> str:
        return self._memory_request

    @memory_request.setter
    def memory_request(self, memory_request):
        ResourceRequest._verify_memory_request(memory_request)
        self._memory_request = memory_request

    @classmethod
    def _verify_min_max_replica(cls, min_replica, max_replica):
        if min_replica < ResourceRequest.min_allowed_replica:
            raise InvalidReplicaCountException(
                f"Min replica count must be >= {ResourceRequest.min_allowed_replica}; "
                f"min_replica passed: {min_replica}")
        elif max_replica > ResourceRequest.max_allowed_replica:
            raise InvalidReplicaCountException(
                f"Max replica count must be <= {ResourceRequest.max_allowed_replica}; "
                f"min_replica passed: {max_replica}")
        elif min_replica >= max_replica:
            raise InvalidReplicaCountException(
                f"Min replica must be < max_replica; "
                f"min_replica passed: {min_replica}, max_replica passed: {max_replica}"
            )

    @classmethod
    def _verify_cpu_request(cls, cpu_request: str):
        matched = re.fullmatch(r"^(\d{1,3}(\.\d{1,3})?)$|^(\d{2,5}m)$", cpu_request)
        if bool(matched) is False:
            raise InvalidCPURequestException(
                f'Valid CPU value is required, e.g "2" or "500m"; cpu_request passed: {cpu_request}'
            )

    @classmethod
    def _verify_memory_request(cls, memory_request: str):
        matched = re.fullmatch(r"^\d+(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?)?$", memory_request)
        if bool(matched) is False:
            raise InvalidMemoryRequestException(
                f"Valid RAM value is required, e.g. 512Mi; memory_request passed: {memory_request}"
            )

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.ResourceRequest(
            min_replica=self.min_replica,
            max_replica=self.max_replica,
            cpu_request=self.cpu_request,
            memory_request=self.memory_request
        )


class InvalidReplicaCountException(Exception):
    pass


class InvalidCPURequestException(Exception):
    pass


class InvalidMemoryRequestException(Exception):
    pass
