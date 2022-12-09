import { get } from "../../components/form/utils";
import { Ensembler, NopEnsembler } from "../ensembler";
import {
  BaseExperimentEngine,
  NopExperimentEngine,
} from "../experiment_engine";
import { Status } from "../status/Status";

const _ = require(`lodash`);
const objectAssignDeep = require(`object-assign-deep`);

export class TuringRouter {
  constructor() {
    this.id = 0;
    this.environment_name = "";
    this.name = "";
    this.status = undefined;
    this.config = {
      routes: [newRoute()],
      default_route_id: null,
      default_traffic_rule: null,
      rules: [],
      resource_request: {
        cpu_request: "500m",
        memory_request: "512Mi",
        min_replica: 0,
        max_replica: 2,
      },
      autoscaling_policy: {
        metric: null,
        target: null,
        payload_size: "200Mi"
      },
      timeout: "300ms",
      protocol: "HTTP_JSON",
      experiment_engine: new NopExperimentEngine(),
      enricher: {
        type: "nop",
        timeout: "100ms",
        endpoint: "/",
        port: 8080,
        resource_request: {
          cpu_request: "500m",
          memory_request: "512Mi",
          min_replica: 0,
          max_replica: 2,
        },
        autoscaling_policy: {
          metric: null,
          target: null,
          payload_size: "200Mi"
        },
        env: [],
        service_account: "",
      },
      ensembler: new NopEnsembler(),
      log_config: {
        result_logger_type: "nop",
        bigquery_config: {
          table: "",
          service_account_secret: "",
        },
        kafka_config: {
          brokers: "",
          topic: "",
          serialization_format: "protobuf",
        },
      },
    };
  }

  static fromJson(json) {
    const router = objectAssignDeep(new TuringRouter(), json);
    router.status = Status.fromValue(json.status);
    // If the router has just been created, there is no config while it's being deployed.
    // Clear the dummy config.
    if (!json.config) {
      router.config = undefined;
      return router;
    }

    // Init experiment engine
    router.config.experiment_engine = BaseExperimentEngine.fromJson(
      get(json, "config.experiment_engine")
    );

    // Init ensembler. If type nop / standard, send in the default route id.
    const ensemblerConfig = get(json, "config.ensembler");
    router.config.ensembler = _.isEmpty(ensemblerConfig)
      ? Ensembler.fromJson({
          nop_config: {
            final_response_route_id: get(json, "config.default_route_id"),
          },
        })
      : ensemblerConfig.type === "standard"
      ? Ensembler.fromJson({
          ...ensemblerConfig,
          standard_config: {
            ...ensemblerConfig.standard_config,
            fallback_response_route_id: get(json, "config.default_route_id"),
          },
        })
      : Ensembler.fromJson(ensemblerConfig);

    // Init enricher. If config exists, update the type to docker.
    const enricherConfig = get(json, "config.enricher");
    if (!!enricherConfig && enricherConfig.type !== "nop") {
      router.config.enricher = { ...router.config.enricher, type: "docker" };
    }

    return router;
  }

  toJSON() {
    let obj = objectAssignDeep({}, this);

    // Remove properties for optional fields, if not relevant
    // Default Traffic Rule
    if (!obj.config.default_traffic_rule) {
      delete obj.config["default_traffic_rule"];
    }

    // If the default autoscaling policy is used for the router, there is no need for the resource requests field
    if (obj.config.autoscaling_policy.payload_size !== "") {
      delete obj.config["resource_request"]
    }

    // Enricher
    if (obj.config.enricher && obj.config.enricher.type === "nop") {
      delete obj.config["enricher"];
    } else {
      // If the default autoscaling policy is used for the enricher, there is no need for the resource requests field
      if (obj.config.enricher.autoscaling_policy.payload_size !== "") {
        delete obj.config.enricher["resource_request"]
      }
    }

    // Ensembler
    if (obj.config.ensembler.type === "nop") {
      // Copy the final response route id to the top level, as the default route
      obj.config.default_route_id =
        obj.config["ensembler"].nop_config["final_response_route_id"];
      delete obj.config["ensembler"];
    } else if (obj.config.ensembler.type === "standard") {
      // Copy the fallback response route id to the top level, as the default route
      obj.config.default_route_id =
        obj.config["ensembler"].standard_config["fallback_response_route_id"];
      delete obj.config["ensembler"].standard_config[
        "fallback_response_route_id"
      ];
    } else {
      // Docker or Pyfunc ensembler, clear the default_route_id
      delete obj.config["default_route_id"];
      // If the default autoscaling policy is used for the ensembler, there is no need for the resource requests field
      if (obj.config?.ensembler?.docker_config.autoscaling_policy.payload_size !== "") {
        delete obj.config.ensembler.docker_config["resource_request"]
      }
      if (obj.config.ensembler.type === "pyfunc") {
        // If the default autoscaling policy is used for the ensembler, there is no need for the resource requests field
        if (obj.config.ensembler.pyfunc_config.autoscaling_policy.payload_size !== "") {
          delete obj.config.ensembler.pyfunc_config["resource_request"]
        }
        // Delete the docker config
        delete obj.config["ensembler"].docker_config;
      }
    }

    // Outcome Logging
    if (
      obj.config.log_config.bigquery_config &&
      obj.config.log_config.result_logger_type !== "bigquery"
    ) {
      delete obj.config.log_config["bigquery_config"];
    }
    if (
      obj.config.log_config.kafka_config &&
      obj.config.log_config.result_logger_type !== "kafka"
    ) {
      delete obj.config.log_config["kafka_config"];
    }
    return obj;
  }
}

export const newRoute = (protocol) => ({
  id: "",
  type: "PROXY",
  timeout: "100ms",
  service_method:
    protocol === "UPI_V1"
      ? "/caraml.upi.v1.UniversalPredictionService/PredictValues"
      : "",
});

export const newDefaultRule = () => ({
  routes: [],
});

export const newRule = () => ({
  name: "",
  conditions: [],
  routes: [],
});
