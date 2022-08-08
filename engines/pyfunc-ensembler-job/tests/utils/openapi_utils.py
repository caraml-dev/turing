import yaml
from turing.generated import Configuration
from turing.generated.model_utils import validate_and_convert_types


def from_yaml(text: str, required_type):
    json_dict = yaml.safe_load(text)
    return validate_and_convert_types(
        input_value=json_dict,
        required_types_mixed=(required_type,),
        path_to_item=[],
        spec_property_naming=True,
        _check_type=True,
        configuration=Configuration(),
    )
