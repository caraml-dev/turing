from typing import TypeVar, SupportsAbs
import google.protobuf.message
import google.protobuf.json_format
import yaml

T = TypeVar('T', bound=SupportsAbs[google.protobuf.message.Message])


def from_yaml(text: str, message: 'T') -> 'T':
    json_dict = yaml.safe_load(text)
    return google.protobuf.json_format.ParseDict(json_dict, message)
