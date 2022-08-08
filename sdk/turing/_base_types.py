from datetime import datetime
from turing.generated.model_utils import OpenApiModel


class DataObject:
    """
    Base class to standardize the conversion of inherited classes into a generic
    JSON-like dictionary objects
    """

    def to_dict(self):
        attribs = [(k, v) for k, v in self.__dict__.items() if not k.startswith("_")]

        for name in dir(self.__class__):
            # skip private properties
            if name.startswith("_"):
                continue
            obj = getattr(self.__class__, name)
            if isinstance(obj, property):
                val = obj.__get__(self, self.__class__)
                if val:
                    attribs.append((name, val))

        return dict(attribs)

    def __eq__(self, other):
        if isinstance(self, other.__class__):
            return self.to_dict() == other.to_dict()
        return False

    def __str__(self):
        return "%s(%s)" % (
            type(self).__name__,
            ",".join("\n\t%s=%s" % item for item in self.to_dict().items()),
        )

    def __repr__(self) -> str:
        return self.__str__()


class ApiObject(DataObject):
    """
    Base DTO class, for objects retrieved from Turing API
    """

    def __init__(
        self,
        id: int = 0,
        created_at: datetime = None,
        updated_at: datetime = None,
        **kwargs
    ):
        self._id = id
        self._created_at = created_at
        self._updated_at = updated_at

    @property
    def id(self) -> int:
        return self._id

    @property
    def created_at(self) -> datetime:
        return self._created_at

    @property
    def updated_at(self) -> datetime:
        return self._updated_at

    @classmethod
    def from_open_api(cls, open_api: OpenApiModel):
        """
        Factory method, for constructing ApiObject (and its sub-classes) instances
        from a relevant instance of openapi-generator generated model instance

        :param open_api: instance of openapi-generator model
        :return: new ApiObject instance
        """
        return cls(**open_api.to_dict())

    def to_open_api(self):
        """
        Converts this ApiObject into an instance of openapi-generator model.
        Sub-class of ApiObject should either
        - have _OPEN_API_SPEC field set to a respective type from turing.generated.model model class
          Example:

          class Project(ApiObject):
            _OPEN_API_SPEC = turing.generated.models.Project
        or
        - be annotated with ApiObjectSpec annotation
          Example:

          @ApiObjectSpec(turing.generated.models.Project)
          class Project(ApiObject):

        :return: instance of respective openapi data-transfer object
        """
        return self._OPEN_API_SPEC(**self.to_dict())


def ApiObjectSpec(spec: object):
    def wrap(cls):
        cls._OPEN_API_SPEC = spec
        return cls

    return wrap
