from datetime import datetime

import json

from turing.generated.model_utils import ModelNormal


class ApiObject:

    def __init__(
            self,
            id: int = 0,
            created_at: datetime = None,
            updated_at: datetime = None,
            **kwargs):
        self._id = id
        self._created_at = created_at
        self._updated_at = updated_at

    @property
    def id(self) -> int:
        return self._id

    @id.setter
    def id(self, id: int):
        self._id = id

    @property
    def created_at(self) -> datetime:
        return self._created_at

    @created_at.setter
    def created_at(self, created_at: datetime):
        self._created_at = created_at

    @property
    def updated_at(self) -> datetime:
        return self._updated_at

    @updated_at.setter
    def updated_at(self, updated_at: datetime):
        self._updated_at = updated_at

    @classmethod
    def from_open_api(cls, open_api: ModelNormal):
        return cls(**open_api.to_dict())

    def to_open_api(self):
        attribs = [(k, v) for k, v in self.__dict__.items() if not k.startswith("_")]

        for name in dir(self.__class__):
            # a protected property is somewhat uncommon but
            # let's stay consistent with plain attribs
            if name.startswith("_"):
                continue
            obj = getattr(self.__class__, name)
            if isinstance(obj, property):
                val = obj.__get__(self, self.__class__)
                if val:
                    attribs.append((name, val))

        return self.OPEN_API_SPEC(**dict(attribs))


def ApiObjectSpec(spec: object):
    def wrap(cls):
        cls.OPEN_API_SPEC = spec
        return cls

    return wrap
