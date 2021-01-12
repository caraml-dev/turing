import yaml from "js-yaml";
import objectAssignDeep from "object-assign-deep";

export class RouterVersion {
  static fromJson(json) {
    return objectAssignDeep(new RouterVersion(), json);
  }

  toPrettyYaml() {
    const pretty = {
      version: this.version,
      router: {
        image: this.image,
        timeout: this.timeout,
        routes: this.routes.map(route => ({
          id: route.id,
          endpoint: route.endpoint,
          timeout: route.timeout,
          is_default: route.id === this.default_route_id
        })),
        resource_request: this.resource_request
      },
      experiment_engine: {
        type: this.experiment_engine.type,
        ...(this.experiment_engine.type !== "nop"
          ? {
              config: {
                client: {
                  id: this.experiment_engine.config.client.username,
                  encrypted_passkey: this.experiment_engine.config.client.passkey
                },
                experiments: this.experiment_engine.config.experiments,
                variables: this.experiment_engine.config.variables
              }
            }
          : undefined)
      },
      enricher: !!this.enricher
        ? {
            type: "docker",
            image: this.enricher.image,
            endpoint: this.enricher.endpoint,
            port: this.enricher.port,
            env: this.enricher.env,
            service_account: this.enricher.service_account,
            resource_request: this.enricher.resource_request
          }
        : {
            type: "none"
          },
      ensembler: !!this.ensembler
        ? {
            type: this.ensembler.type,
            ...(this.ensembler.type === "standard"
              ? {
                  standard_config: this.ensembler.standard_config
                }
              : this.ensembler.type === "docker"
              ? {
                  image: this.ensembler.image,
                  endpoint: this.ensembler.endpoint,
                  port: this.ensembler.port,
                  env: this.ensembler.env,
                  service_account: this.ensembler.service_account,
                  resource_request: this.ensembler.resource_request
                }
              : undefined)
          }
        : {
            type: "none"
          },
      result_logging: {
        type: this.log_config.result_logger_type,
        ...(this.log_config.result_logger_type === "bigquery"
          ? {
              config: {
                table: this.log_config.bigquery_config.table,
                service_account: this.log_config.bigquery_config
                  .service_account_secret
              }
            }
          : this.log_config.result_logger_type === "kafka"
          ? {
              config: this.log_config.kafka_config
            }
          : undefined)
      }
    };

    return yaml.dump(pretty);
  }
}
