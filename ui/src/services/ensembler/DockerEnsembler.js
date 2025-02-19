import { Ensembler } from "./Ensembler";

const objectAssignDeep = require(`object-assign-deep`);

export class DockerEnsembler extends Ensembler {
  constructor() {
    super("docker");
    this.docker_config = DockerEnsembler.newConfig();

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json = {}) {
    const ensembler = new DockerEnsembler();
    ensembler.docker_config = objectAssignDeep({}, json.docker_config);
    return ensembler;
  }

  static newConfig() {
    return {
      timeout: "100ms",
      endpoint: "/",
      port: 8080,
      resource_request: {
        cpu_request: "500m",
        cpu_limit: "",
        memory_request: "512Mi",
        min_replica: 0,
        max_replica: 2,
      },
      autoscaling_policy: {
        metric: "concurrency",
        target: "1",
      },
      env: [],
      secrets: [],
      service_account: "",
    };
  }
}
