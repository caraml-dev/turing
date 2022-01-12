import re
from abc import ABC, abstractmethod


class Schema(ABC):
    regex_exp = None

    @classmethod
    @abstractmethod
    def verify_schema(cls, value):
        pass


class DockerImageSchema(Schema):
    regex_exp = r"^([a-z0-9]+(?:[._-][a-z0-9]+)*(?::\d{2,5})?\/)?([a-z0-9]+(?:[._-][a-z0-9]+)*\/)*([a-z0-9]+(?:[._-][a-z0-9]+)*)(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$"

    @classmethod
    def verify_schema(cls, value):
        matched = re.fullmatch(
            DockerImageSchema.regex_exp,
            value,
            re.IGNORECASE
        )
        if bool(matched) is False:
            raise InvalidImageException(
                f"Valid Docker Image value should be provided, e.g. kennethreitz/httpbin:latest; "
                f"image passed: {value}"
            )


class EnvVarNameSchema(Schema):
    regex_exp = r"^[a-z0-9_]*$"

    @classmethod
    def verify_schema(cls, value):
        matched = re.fullmatch(EnvVarNameSchema.regex_exp, value, re.IGNORECASE)
        if bool(matched) is False:
            raise InvalidEnvironmentVariableNameException(
                f"The name of a variable can contain only alphanumeric character or the underscore; "
                f"name passed: {value}"
            )


class TimeoutSchema(Schema):
    regex_exp = r"^[0-9]+(ms|s|m|h)$"

    @classmethod
    def verify_schema(cls, value):
        matched = re.fullmatch(TimeoutSchema.regex_exp, value)
        if bool(matched) is False:
            raise InvalidTimeoutException(
                f"Valid duration is required; timeout passed: {value}"
            )


class CpuRequestSchema(Schema):
    regex_exp = r"^(\d{1,3}(\.\d{1,3})?)$|^(\d{2,5}m)$"

    @classmethod
    def verify_schema(cls, value):
        matched = re.fullmatch(CpuRequestSchema.regex_exp, value)
        if bool(matched) is False:
            raise InvalidCPURequestException(
                f'Valid CPU value is required, e.g "2" or "500m"; cpu_request passed: {value}'
            )


class MemoryRequestSchema(Schema):
    regex_exp = r"^\d+(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?)?$"

    @classmethod
    def verify_schema(cls, value):
        matched = re.fullmatch(MemoryRequestSchema.regex_exp, value)
        if bool(matched) is False:
            raise InvalidMemoryRequestException(
                f"Valid RAM value is required, e.g. 512Mi; memory_request passed: {value}"
            )


class InvalidImageException(Exception):
    pass


class InvalidEnvironmentVariableNameException(Exception):
    pass


class InvalidTimeoutException(Exception):
    pass


class InvalidCPURequestException(Exception):
    pass


class InvalidMemoryRequestException(Exception):
    pass

