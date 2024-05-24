import { Ensembler } from "./Ensembler";

const objectAssignDeep = require(`object-assign-deep`);

export class PyFuncEnsembler extends Ensembler {
  constructor() {
    super("pyfunc");
    this.pyfunc_config = PyFuncEnsembler.newConfig();

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json = {}) {
    const ensembler = new PyFuncEnsembler();
    ensembler.pyfunc_config = objectAssignDeep({}, json.pyfunc_config);
    if (!!json.docker_config) {
      ensembler.docker_config = objectAssignDeep({}, json.docker_config);
    }
    return ensembler;
  }

  static newConfig(project_id) {
    return {
      project_id: project_id,
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
      timeout: "100ms",
    };
  }
}
