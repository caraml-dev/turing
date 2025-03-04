/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";
import { get } from "@caraml-dev/ui-lib";
import { standardExperimentConfigSchema } from "../components/experiment_config/validation/schema";
import {
  fieldSchema,
  fieldSourceSchema,
} from "../../../../components/validation";
import { autoscalingPolicyMetrics } from "../components/autoscaling_policy/typeOptions";

yup.addMethod(yup.array, "unique", function(propertyPath, message) {
  return this.test("unique", message, function(list) {
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

const validateDanglingRoutes = function(items) {
  const defaultTrafficRule = this.options.parent.default_traffic_rule;
  let trafficRuleRoutes = [
    ...new Set(this.options.parent.rules.map((rule) => rule.routes).flat(1)),
  ];
  if (defaultTrafficRule) {
    trafficRuleRoutes = [...trafficRuleRoutes, ...defaultTrafficRule.routes];
  }

  const errors = [];
  items.forEach((item, idx) => {
    if (!trafficRuleRoutes.includes(item.id)) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].id`,
          message:
            "This route should be removed since it has no Traffic Rule(s) associated and will never be called.",
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
  configMapNameRegex = /^[-._a-zA-Z0-9]+$/,
  bigqueryTableRegex = /^[a-z][a-z0-9-]+\.\w+([_]?\w)+\.\w+([_]?\w)+$/i,
  kafkaBrokersRegex =
    /^([a-z]+:\/\/)?\[?([0-9a-zA-Z\-%._:]*)\]?:([0-9]+)(,([a-z]+:\/\/)?\[?([0-9a-zA-Z\-%._:]*)\]?:([0-9]+))*$/i,
  kafkaTopicRegex = /^[A-Za-z0-9_.-]{1,249}/i,
  trafficRuleNameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/,
  upiUrlRegex = /^([a-zA-Z0-9_-]{1,256})(\.[a-zA-Z0-9-]{1,256})*(:\d+)$/,
  serviceMethodRegex = /^\/?(\w+(\.|-)?)+(\/[A-Za-z0-9]+)$/;

const timeoutSchema = yup
  .string()
  .matches(durationRegex, "Valid duration is required");

const routeSchema = yup.object().shape({
  id: yup.string().required("Valid route Id is required"),
  type: yup.string().oneOf(["PROXY"], "Route Type is required"),
  endpoint: yup.string().when("$protocol", {
    is: "UPI_V1",
    then: (schema) =>
      schema
        .required("Valid gRPC endpoint is required eg. {host}:{port}")
        .matches(
          upiUrlRegex,
          "Valid gRPC endpoint is required eg. {host}:{port}"
        ),
    otherwise: (schema) =>
      schema.required("Valid url is required").url("Valid url is required"),
  }),
  timeout: timeoutSchema.required("Timeout is required"),
  service_method: yup.string().when("$protocol", {
    is: "UPI_V1",
    then: (schema) =>
      schema.matches(
        serviceMethodRegex,
        "Valid service method is required in format {package}/{method} eg. my.package/PredictValues"
      ),
  }),
});

const validRouteSchema = yup
  .string()
  .test("valid-route", "Valid route is required", function(value) {
    const configSchema = this.from.slice(-1).pop();
    const { routes } = configSchema.value.config;
    return routes.map((r) => r.id).includes(value);
  });

const ruleConditionSchema = yup.object().shape({
  field_source: fieldSourceSchema(),
  field: fieldSchema("field_source")(),
  operator: yup
    .string()
    .oneOf(["in"], "One of supported operators should be specified"),
  values: yup
    .array(yup.string())
    .min(1, "At least one value should be provided"),
});

const defaultTrafficRuleSchema = (_) => yup.object().shape({
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
    .test(
      "is-not-default-name",
      "default-traffic-rule is a reserved name, and cannot be used as the name for a Custom Traffic Rule.",
      (value) => value !== "default-traffic-rule"
    ),
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
      "The name of an environment variable must contain only alphanumeric characters or '_'"
    ),
  value: yup.string(),
});

const secretSchema = yup.object().shape({
  mlp_secret_name: yup
    .string()
    .required("MLP secret name is required")
    .matches(
      configMapNameRegex,
      "The name of the MLP secret must contain only alphanumeric characters, '-', '_' or '.'"
    ),
  env_var_name: yup
    .string()
    .required("Environment variable name is required")
    .matches(
      envVariableNameRegex,
      "The name of an environment variable must contain only alphanumeric characters or '_'"
    ),
});

const resourceRequestSchema = (maxAllowedReplica) =>
  yup.object().shape({
    cpu_request: yup
      .string()
      .matches(
        cpuRequestRegex,
        'Valid CPU value is required, e.g "2" or "500m"'
      ),
    cpu_limit: yup
      .string()
      .matches(
        cpuRequestRegex,
        { message: 'Valid CPU value is required, e.g "2" or "500m"', excludeEmptyString: true }
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
      .when("min_replica", ([minReplica], schema) =>
        minReplica === 0
          ? schema.positive("Max Replica should be positive")
          : schema
      ),
  });

const autoscalingPolicySchema = yup.object().shape({
  metric: yup
    .string()
    .required("Valid metric must be chosen")
    .oneOf(
      autoscalingPolicyMetrics,
      "Valid autoscaling metric type should be chosen"
    ),
  target: yup
    .string()
    .required("Valid target should be specified")
    .matches(/(\d+(?:\.\d+)?)/, "Must be a number"),
});

const enricherSchema = yup.object().shape({
  type: yup
    .string()
    .required("Valid Enricher type should be selected")
    .oneOf(["nop", "docker"], "Valid Enricher type should be selected"),
});

const dockerImageSchema = yup
  .string()
  .matches(
    dockerImageRegex,
    "Valid Docker Image value should be provided, e.g. kennethreitz/httpbin:latest"
  );

const dockerDeploymentSchema = (maxAllowedReplica) => (_) =>
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
    secrets: yup.array(secretSchema),
    resource_request: resourceRequestSchema(maxAllowedReplica),
    autoscaling_policy: autoscalingPolicySchema,
  });

const pyfuncDeploymentSchema = (maxAllowedReplica) => (_) =>
  yup.object().shape({
    project_id: yup.number().integer().required("Project ID is required"),
    ensembler_id: yup.number().integer().required("Ensembler ID is required"),
    timeout: timeoutSchema.required("Timeout is required"),
    resource_request: resourceRequestSchema(maxAllowedReplica),
    autoscaling_policy: autoscalingPolicySchema,
    env: yup.array(environmentVariableSchema),
    secrets: yup.array(secretSchema),
  });

const mappingSchema = yup.object().shape({
  experiment: yup.string().required("Experiment name is required"),
  treatment: yup.string().required("Treatment name is required"),
  route: yup.string().required("Treatment needs to be mapped back to a route"),
});

const standardEnsemblerConfigSchema = (_) => yup
  .object()
  .shape({
    route_name_path: yup.string().nullable(),
    experiment_mappings: yup.array(mappingSchema).nullable(),
    fallback_response_route_id: validRouteSchema,
    lazy_routing: yup.bool(),
  })
  .test(
    "is-route-name-path-or-experiment-mappings-set",
    "only one of route_name_path or experiment_mappings must be set",
    (standardEnsembler) => {
      const isRouteNamePathEmpty = !standardEnsembler.route_name_path;
      const isExperimentMappingsEmpty =
        !standardEnsembler.experiment_mappings ||
        standardEnsembler.experiment_mappings.length === 0;
      return !(
        (isRouteNamePathEmpty && isExperimentMappingsEmpty) ||
        (!isRouteNamePathEmpty && !isExperimentMappingsEmpty)
      );
    }
  );

const bigQueryConfigSchema = (_) => yup.object().shape({
  table: yup
    .string()
    .required("BigQuery table name is required")
    .matches(
      bigqueryTableRegex,
      "Valid BQ table name is required, e.g. project_name.dataset.table"
    ),
  service_account_secret: yup.string().required("Service Account is required"),
});

const kafkaConfigSchema = (_) => yup.object().shape({
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
    .string()
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
      rules: yup
        .array(trafficRuleSchema)
        .test("unique-rule-names", validateRuleNames),
      routes: yup
        .array(routeSchema)
        .unique("id", "Route Id must be unique")
        .min(1, "At least one route should be configured")
        .when(["rules"], ([rules], schema) => {
          if (rules.length > 0) {
            return schema.test("no-dangling-routes", validateDanglingRoutes);
          }
        }),
      default_traffic_rule: yup
        .object()
        .nullable()
        .when("rules", {
          is: (rules) => rules.length > 0,
          then: defaultTrafficRuleSchema,
        }),
      resource_request: resourceRequestSchema(maxAllowedReplica),
      autoscaling_policy: autoscalingPolicySchema,
      protocol: yup
        .string()
        .required("Valid Protocol should be selected")
        .oneOf(["HTTP_JSON", "UPI_V1"], "Valid Protocol should be selected"),
    }),
  }),
  yup.object().shape({
    config: yup.object().shape({
      experiment_engine: yup.object().shape({
        type: yup
          .string()
          .required("Valid Experiment Engine should be selected")
          .when("$experimentEngineOptions", ([options], schema) =>
            schema.oneOf(options, "Valid Experiment Engine should be selected")
          ),
        config: yup.object().when("type", ([engine], schema) =>
          engine === "nop"
            ? schema
            : schema
              .when("$getEngineProperties", ([getEngineProperties], schema) => {
                const engineProps = getEngineProperties(engine);
                return engineProps.type === "standard"
                  ? standardExperimentConfigSchema(engineProps)(schema)
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
              dockerDeploymentSchema(maxAllowedReplica)()
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
          .string()
          .required("Valid Ensembler type should be selected")
          .oneOf(
            ["nop", "docker", "standard", "pyfunc"],
            "Valid Ensembler type should be selected"
          ),
        nop_config: yup.object().when("type", {
          is: "nop",
          then: (_) => yup.object().shape({
            final_response_route_id: validRouteSchema,
          }),
        }),
        some_field: yup.string().when("some_other_field", ([some_other_field], schema) =>
          some_other_field ? yup.string().required("this is needed") : schema
        ),
        docker_config: yup.object().when("type", {
          is: "docker",
          then: dockerDeploymentSchema(maxAllowedReplica),
        }),
        standard_config: yup.object().when("type", {
          is: "standard",
          then: standardEnsemblerConfigSchema,
        }),
        pyfunc_config: yup.object().when("type", {
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
          .string()
          .required("Valid Results Logging type should be selected")
          .oneOf(
            ["nop", "bigquery", "kafka"],
            "Valid Results Logging type should be selected"
          ),
        bigquery_config: yup.object().when("result_logger_type", {
          is: "bigquery",
          then: bigQueryConfigSchema,
        }),
        kafka_config: yup.object().when("result_logger_type", {
          is: "kafka",
          then: kafkaConfigSchema,
        }),
      }),
    }),
  }),
];

export default schema;
