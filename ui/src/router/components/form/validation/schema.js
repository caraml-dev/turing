/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";
import { get } from "@gojek/mlp-ui";
import { standardExperimentConfigSchema } from "../components/experiment_config/validation/schema";
import {
  fieldSchema,
  fieldSourceSchema,
} from "../../../../components/validation";

yup.addMethod(yup.array, "unique", function (propertyPath, message) {
  return this.test("unique", message, function (list) {
    const errors = [];

    list.forEach((item, index) => {
      const propertyValue = get(item, propertyPath);

      if (
        propertyValue &&
        list.some(
          (other) =>
            other !== item && get(other, propertyPath) === propertyValue
        )
      ) {
        errors.push(
          this.createError({
            path: `${this.path}[${index}].${propertyPath}`,
            message,
          })
        );
      }
    });

    if (errors.length > 0) {
      throw new yup.ValidationError(errors);
    }

    return true;
  });
});

const validateRuleNames = function(items) {
  const uniqueNamesMap = items.reduce((acc, item) => {
    const current = item.name in acc ? acc[item.name] : 0;
    // If name is set, increment the count
    return !!item.name ? { ...acc, [item.name]: current + 1 } : acc;
  }, {});
  const errors = [];
  items.forEach((item, idx) => {
    if (!!item.name && uniqueNamesMap[item.name] > 1) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].name`,
          message: "Rule names in a Router should be unique",
        })
      );
    }
  });
  return !!errors.length ? new yup.ValidationError(errors) : true;
};

const routerNameRegex = /^[a-z0-9-]*$/,
  durationRegex = /^[0-9]+(ms|s|m|h)$/,
  cpuRequestRegex = /^(\d{1,3}(\.\d{1,3})?)$|^(\d{2,5}m)$/,
  memRequestRegex = /^\d+(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?)?$/,
  envVariableNameRegex = /^[a-z0-9_]*$/i,
  dockerImageRegex =
    /^([a-z0-9]+(?:[._-][a-z0-9]+)*(?::\d{2,5})?\/)?([a-z0-9]+(?:[._-][a-z0-9]+)*\/)*([a-z0-9]+(?:[._-][a-z0-9]+)*)(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$/i,
  bigqueryTableRegex = /^[a-z][a-z0-9-]+\.\w+([_]?\w)+\.\w+([_]?\w)+$/i,
  kafkaBrokersRegex =
    /^([a-z]+:\/\/)?\[?([0-9a-zA-Z\-%._:]*)\]?:([0-9]+)(,([a-z]+:\/\/)?\[?([0-9a-zA-Z\-%._:]*)\]?:([0-9]+))*$/i,
  kafkaTopicRegex = /^[A-Za-z0-9_.-]{1,249}/i,
  trafficRuleNameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/;

const timeoutSchema = yup
  .string()
  .matches(durationRegex, "Valid duration is required");

const routeSchema = yup.object().shape({
  id: yup.string().required("Valid route Id is required"),
  type: yup.string().oneOf(["PROXY"], "Route Type is required"),
  endpoint: yup
    .string()
    .required("Valid url is required")
    .url("Valid url is required"),
  timeout: timeoutSchema.required("Timeout is required"),
});

const validRouteSchema = yup
  .mixed()
  .test("valid-route", "Valid route is required", function (value) {
    const configSchema = this.from.slice(-1).pop();
    const { routes } = configSchema.value.config;
    return routes.map((r) => r.id).includes(value);
  });

const ruleConditionSchema = yup.object().shape({
  field_source: fieldSourceSchema,
  field: fieldSchema("field_source"),
  operator: yup
    .mixed()
    .oneOf(["in"], "One of supported operators should be specified"),
  values: yup
    .array(yup.string())
    .required("At least one value should be provided"),
});

const defaultTrafficRuleSchema = yup.object().shape({
  routes: yup
    .array()
    .of(validRouteSchema)
    .min(1, "At least one route should be attached to the rule"),
});

const trafficRuleSchema = yup.object().shape({
  name: yup
    .string()
    .required("Rule name is required")
    .matches(
      trafficRuleNameRegex,
      "Name must begin with an alphanumeric character and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$%&:."
    )
    .test('is-not-default-name', "default-traffic-rule is a reserved name, and cannot be used as the name for a Custom Traffic Rule.", (value) => value !== "default-traffic-rule"),
  conditions: yup
    .array()
    .of(ruleConditionSchema)
    .min(1, "At least one condition should be defined"),
  routes: yup
    .array()
    .of(validRouteSchema)
    .min(1, "At least one route should be attached to the rule"),
});

const environmentVariableSchema = yup.object().shape({
  name: yup
    .string()
    .required("Variable name can not be empty")
    .matches(
      envVariableNameRegex,
      "The name of a variable can contain only alphanumeric character or the underscore"
    ),
  value: yup.string(),
});

const resourceRequestSchema = (maxAllowedReplica) =>
  yup.object().shape({
    cpu_request: yup
      .string()
      .matches(
        cpuRequestRegex,
        'Valid CPU value is required, e.g "2" or "500m"'
      ),
    memory_request: yup
      .string()
      .matches(memRequestRegex, "Valid RAM value is required, e.g. 512Mi"),
    min_replica: yup
      .number()
      .typeError("Min Replicas value is required")
      .min(0, "Min Replicas can not be less than 0"),
    max_replica: yup
      .number()
      .typeError("Max Replicas value is required")
      .min(
        yup.ref(`min_replica`),
        "Max Replicas can not be less than Min Replicas"
      )
      .max(
        maxAllowedReplica,
        "Max Replicas value has exceeded allowed number of replicas: ${max}"
      )
      .when("min_replica", (minReplica, schema) =>
        minReplica === 0
          ? schema.positive("Max Replica should be positive")
          : schema
      ),
  });

const enricherSchema = yup.object().shape({
  type: yup
    .mixed()
    .required("Valid Enricher type should be selected")
    .oneOf(["nop", "docker"], "Valid Enricher type should be selected"),
});

const dockerImageSchema = yup
  .string()
  .matches(
    dockerImageRegex,
    "Valid Docker Image value should be provided, e.g. kennethreitz/httpbin:latest"
  );

const dockerDeploymentSchema = (maxAllowedReplica) =>
  yup.object().shape({
    image: dockerImageSchema.required("Docker Image is required"),
    endpoint: yup.string().required("Endpoint value is required"),
    port: yup
      .number()
      .integer()
      .typeError("Port value is required, e.g. 8080")
      .required("Port value is required, e.g. 8080"),
    timeout: timeoutSchema.required("Timeout is required"),
    env: yup.array(environmentVariableSchema),
    resource_request: resourceRequestSchema(maxAllowedReplica),
  });

const pyfuncDeploymentSchema = (maxAllowedReplica) =>
  yup.object().shape({
    project_id: yup.number().integer().required("Project ID is required"),
    ensembler_id: yup.number().integer().required("Ensembler ID is required"),
    timeout: timeoutSchema.required("Timeout is required"),
    resource_request: resourceRequestSchema(maxAllowedReplica),
    env: yup.array(environmentVariableSchema),
  });

const mappingSchema = yup.object().shape({
  experiment: yup.string().required("Experiment name is required"),
  treatment: yup.string().required("Treatment name is required"),
  route: yup.string().required("Treatment needs to be mapped back to a route"),
});

const standardEnsemblerConfigSchema = yup.object().shape({
  experiment_mappings: yup.array(mappingSchema),
  fallback_response_route_id: validRouteSchema,
});

const bigQueryConfigSchema = yup.object().shape({
  table: yup
    .string()
    .required("BigQuery table name is required")
    .matches(
      bigqueryTableRegex,
      "Valid BQ table name is required, e.g. project_name.dataset.table"
    ),
  service_account_secret: yup.string().required("Service Account is required"),
});

const kafkaConfigSchema = yup.object().shape({
  brokers: yup
    .string()
    .required("Kafka broker(s) is required")
    .matches(
      kafkaBrokersRegex,
      "One or more valid Kafka brokers is required, e.g. host1:port1,host2:port2"
    ),
  topic: yup
    .string()
    .required("Kafka topic name is required")
    .matches(
      kafkaTopicRegex,
      "A valid Kafka topic name may only contain letters, numbers, dot, hyphen or underscore"
    ),
  serialization_format: yup
    .mixed()
    .required("Serialzation format should be selected")
    .oneOf(
      ["json", "protobuf"],
      "Valid serialzation format should be selected"
    ),
});

const schema = (maxAllowedReplica) => [
  yup.object().shape({
    name: yup
      .string()
      .required("Name is required")
      .min(4, "Name should be between 4 and 25 characters")
      .max(25, "Name should be between 4 and 25 characters")
      .matches(
        routerNameRegex,
        "Name can only contain letters a-z (uncapitalized), numbers 0-9 and the dash - symbol"
      ),
    environment_name: yup.string().required("Environment is required"),
    config: yup.object().shape({
      timeout: timeoutSchema.required("Timeout is required"),
      rules: yup.array(trafficRuleSchema).test("unique-rule-names", validateRuleNames),
      routes: yup
        .array(routeSchema)
        .required()
        .unique("id", "Route Id must be unique")
        .min(1, "At least one route should be configured"),
      default_traffic_rule: yup.object()
        .nullable()
        .when('rules', {
          is: rules => rules.length > 0,
          then: defaultTrafficRuleSchema
      }),
      resource_request: resourceRequestSchema(maxAllowedReplica),
    }),
  }),
  yup.object().shape({
    config: yup.object().shape({
      experiment_engine: yup.object().shape({
        type: yup
          .mixed()
          .required("Valid Experiment Engine should be selected")
          .when("$experimentEngineOptions", (options, schema) =>
            schema.oneOf(options, "Valid Experiment Engine should be selected")
          ),
        config: yup.mixed().when("type", (engine, schema) =>
          engine === "nop"
            ? schema
            : yup
                .mixed()
                .when("$getEngineProperties", (getEngineProperties) => {
                  const engineProps = getEngineProperties(engine);
                  return engineProps.type === "standard"
                    ? standardExperimentConfigSchema(engineProps)
                    : engineProps.custom_experiment_manager_config
                        ?.parsed_experiment_config_schema || schema;
                })
        ),
      }),
    }),
  }),
  yup.object().shape({
    config: yup.object().shape({
      enricher: yup.lazy((value) => {
        switch (value.type) {
          case "docker":
            return enricherSchema.concat(
              dockerDeploymentSchema(maxAllowedReplica)
            );
          default:
            return enricherSchema;
        }
      }),
    }),
  }),
  yup.object().shape({
    config: yup.object().shape({
      ensembler: yup.object().shape({
        type: yup
          .mixed()
          .required("Valid Ensembler type should be selected")
          .oneOf(
            ["nop", "docker", "standard", "pyfunc"],
            "Valid Ensembler type should be selected"
          ),
        nop_config: yup.mixed().when("type", {
          is: "nop",
          then: yup.object().shape({
            final_response_route_id: validRouteSchema,
          }),
        }),
        docker_config: yup.mixed().when("type", {
          is: "docker",
          then: dockerDeploymentSchema(maxAllowedReplica),
        }),
        standard_config: yup.mixed().when("type", {
          is: "standard",
          then: standardEnsemblerConfigSchema,
        }),
        pyfunc_config: yup.mixed().when("type", {
          is: "pyfunc",
          then: pyfuncDeploymentSchema(maxAllowedReplica),
        }),
      }),
    }),
  }),
  yup.object().shape({
    config: yup.object().shape({
      log_config: yup.object().shape({
        result_logger_type: yup
          .mixed()
          .required("Valid Results Logging type should be selected")
          .oneOf(
            ["nop", "bigquery", "kafka"],
            "Valid Results Logging type should be selected"
          ),
        bigquery_config: yup.mixed().when("result_logger_type", {
          is: "bigquery",
          then: bigQueryConfigSchema,
        }),
        kafka_config: yup.mixed().when("result_logger_type", {
          is: "kafka",
          then: kafkaConfigSchema,
        }),
      }),
    }),
  }),
];

export default schema;
