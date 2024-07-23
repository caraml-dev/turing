import * as yup from "yup";
import {
  fieldSchema,
  fieldSourceSchema,
} from "../../../../../../components/validation";

const clientSchema = yup.object().shape({
  username: yup.string().required("Client username is required"),
  passkey: yup.string().required("Client passkey is required"),
});

const experimentSchema = yup.object().shape({
  name: yup.string().required("Select one of the experiments available"),
});

const variableSchema = yup.object().shape({
  required: yup.boolean(),
  field_source: yup.string().when("required", {
    is: true,
    then: fieldSourceSchema,
  }),
  field: yup.string().when("required", {
    is: true,
    then: fieldSchema("field_source"),
  }),
});

const standardExperimentConfigSchema = (engineProps) => (_) =>
  yup.object().shape({
    client: engineProps?.standard_experiment_manager_config
      ?.client_selection_enabled
      ? clientSchema
      : yup.object(),
    experiments: engineProps?.standard_experiment_manager_config
      ?.experiment_selection_enabled
      ? yup.array(experimentSchema)
      : yup.array(yup.object()),
    variables: yup.object().shape({
      client_variables: yup.array(),
      experiment_variables: yup.object(),
      config: yup.array(variableSchema),
    }),
  });

export { standardExperimentConfigSchema };
