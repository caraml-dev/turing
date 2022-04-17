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
      rules: [],
      resource_request: {
        cpu_request: "500m",
        memory_request: "512Mi",
        min_replica: 0,
        max_replica: 2,
      },
      timeout: "100ms",
      experiment_engine: new NopExperimentEngine(),
      enricher: {
        type: "nop",
        timeout: "60ms",
        endpoint: "/",
        port: 8080,
        resource_request: {
          cpu_request: "500m",
          memory_request: "512Mi",
          min_replica: 0,
          max_replica: 2,
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
    router.config.experiment_engine = BaseExperimentEngine.fromJson(
      get(json, "config.experiment_engine")
    );

    // Init ensembler. If type nop, send in the default route id.
    const ensemblerConfig = get(json, "config.ensembler");
    router.config.ensembler = _.isEmpty(ensemblerConfig)
      ? Ensembler.fromJson({
          nop_config: {
            final_response_route_id: get(json, "config.default_route_id"),
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
    // Enricher
    if (obj.config.enricher && obj.config.enricher.type === "nop") {
      delete obj.config["enricher"];
    }

    // Ensembler
    if (obj.config.ensembler && obj.config.ensembler.type === "nop") {
      // Copy the final response route id to the top level, as the default route
      obj.config.default_route_id =
        obj.config["ensembler"].nop_config["final_response_route_id"];
      delete obj.config["ensembler"];
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

export const newRoute = () => ({
  id: "",
  type: "PROXY",
  timeout: "20ms",
});

export const newRule = () => ({
  conditions: [],
  routes: [],
});
