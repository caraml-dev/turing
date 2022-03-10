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
    return ensembler;
  }

  static newConfig(project_id) {
    return {
      project_id: project_id,
      resource_request: {
        cpu_request: "500m",
        memory_request: "512Mi",
        min_replica: 0,
        max_replica: 2,
      },
      timeout: "60ms",
    };
  }
}
