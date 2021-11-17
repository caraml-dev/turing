import yaml from "js-yaml";
import objectAssignDeep from "object-assign-deep";
import { BaseExperimentEngine } from "../experiment_engine";
import { get } from "../../components/form/utils";
import { stripKeys } from "../../utils/object";

export class RouterVersion {
  static fromJson(json) {
    const version = objectAssignDeep(new RouterVersion(), json);
    version.experiment_engine = BaseExperimentEngine.fromJson(
      get(json, "experiment_engine")
    );
    return version;
  }

  toPrettyYaml(context) {
    const experiment =
      this.experiment_engine.type === "nop"
        ? {
            type: "none",
          }
        : {
            type: this.experiment_engine.type,
            config:
              context.experiment_engine.type === "custom"
                ? this.experiment_engine.config
                : {
                    client: {
                      id: this.experiment_engine.config.client.username,
                      ...(this.experiment_engine.config.client.passkey_set
                        ? {
                            encrypted_passkey:
                              this.experiment_engine.config.client
                                .encrypted_passkey,
                          }
                        : {
                            encrypted_passkey: "<to be computed>",
                            passkey:
                              this.experiment_engine.config.client.passkey,
                          }),
                    },
                    experiments: this.experiment_engine.config.experiments,
                    variables: stripKeys(
                      this.experiment_engine.config.variables,
                      ["idx", "selected"]
                    ),
                  },
          };

    const pretty = {
      version: this.version,
      router: {
        image: this.image,
        timeout: this.timeout,
        routes: this.routes.map((route) => ({
          id: route.id,
          endpoint: route.endpoint,
          timeout: route.timeout,
          is_default: route.id === this.default_route_id,
        })),
        resource_request: this.resource_request,
      },
      experiment_engine: experiment,
      enricher:
        !!this.enricher && this.enricher.type !== "nop"
          ? {
              type: "docker",
              image: this.enricher.image,
              endpoint: this.enricher.endpoint,
              port: this.enricher.port,
              env: this.enricher.env,
              service_account: this.enricher.service_account,
              resource_request: this.enricher.resource_request,
            }
          : {
              type: "none",
            },
      ensembler:
        !!this.ensembler && this.ensembler.type !== "nop"
          ? {
              type: this.ensembler.type,
              ...(this.ensembler.type === "standard"
                ? {
                    standard_config: this.ensembler.standard_config,
                  }
                : this.ensembler.type === "docker"
                ? {
                    docker_config: this.ensembler.docker_config,
                  }
                : undefined),
            }
          : {
              type: "none",
            },
      result_logging: {
        type: this.log_config.result_logger_type,
        ...(this.log_config.result_logger_type === "bigquery"
          ? {
              config: {
                table: this.log_config.bigquery_config.table,
                service_account:
                  this.log_config.bigquery_config.service_account_secret,
              },
            }
          : this.log_config.result_logger_type === "kafka"
          ? {
              config: this.log_config.kafka_config,
            }
          : undefined),
      },
    };

    return yaml.dump(pretty, { sortKeys: true });
  }
}
