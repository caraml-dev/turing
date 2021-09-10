import * as yup from "yup";
import {
  fieldSchema,
  fieldSourceSchema,
} from "../../../../../../components/validation";

const experimentEngineClientSchema = yup.object().shape({
  username: yup.string().required("Client username is required"),
  passkey: yup.string().required("Client passkey is required"),
});

const experimentEngineExperimentSchema = yup.object().shape({
  name: yup.string().required("Select one of the experiments available"),
});

const experimentEngineVariableSchema = yup.object().shape({
  required: yup.boolean(),
  field_source: yup.mixed().when("required", {
    is: true,
    then: fieldSourceSchema.required(
      "Select the field source for the required variable"
    ),
  }),
  field: yup.mixed().when("required", {
    is: true,
    then: fieldSchema("field_source").required(
      "Specify the field name for the required variable"
    ),
  }),
});

const experimentConfigSchema = yup.object().shape({
  engine: yup.object().shape({
    client_selection_enabled: yup.boolean(),
    experiment_selection_enabled: yup.boolean(),
  }),
  client: yup
    .object()
    .when(["engine.client_selection_enabled"], (clientEnabled, schema) => {
      return clientEnabled ? experimentEngineClientSchema : schema;
    }),
  experiments: yup
    .array()
    .when("engine.experiment_selection_enabled", (experimentEnabled, _) => {
      return experimentEnabled
        ? yup
            .array(experimentEngineExperimentSchema)
            .required("At least one experiment should be configured")
        : yup.array(yup.object());
    }),
  variables: yup.object().shape({
    client_variables: yup.array(),
    experiment_variables: yup.object(),
    config: yup.array(experimentEngineVariableSchema),
  }),
});

export { experimentConfigSchema };
