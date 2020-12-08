import { get } from "../../components/form/utils";
import { Ensembler, NopEnsembler } from "../ensembler";
import {
  BaseExperimentEngine,
  NopExperimentEngine
} from "../experiment_engine";
import { Status } from "../status/Status";

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
        max_replica: 2
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
          max_replica: 2
        },
        env: [],
        service_account: ""
      },
      ensembler: new NopEnsembler(),
      log_config: {
        result_logger_type: "nop",
        bigquery_config: {
          table: "",
          service_account_secret: ""
        },
        kafka_config: {
          brokers: "",
          topic: "",
          serialization_format: ""
        }
      }
    };
  }

  static fromJson(json) {
    const router = objectAssignDeep(new TuringRouter(), json);
    router.status = Status.fromValue(json.status);
    router.config.experiment_engine = BaseExperimentEngine.fromJson(
      get(json, "config.experiment_engine")
    );
    router.config.ensembler = Ensembler.fromJson(get(json, "config.ensembler"));

    const {
      config: { enricher }
    } = router;

    if (!!get(json, "config.enricher")) {
      router.config.enricher = { ...enricher, type: "docker" };
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
  timeout: "20ms"
});

export const newRule = () => ({
  conditions: [],
  routes: []
});
