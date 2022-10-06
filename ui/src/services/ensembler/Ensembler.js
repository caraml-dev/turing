import {
  DockerEnsembler,
  NopEnsembler,
  StandardEnsembler,
  PyFuncEnsembler,
} from "./index";

export class Ensembler {
  constructor(type) {
    this.type = type;
    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json = {}) {
    switch (json.type) {
      case "docker":
        return DockerEnsembler.fromJson(json);
      case "standard":
        return StandardEnsembler.fromJson(json);
      case "pyfunc":
        return PyFuncEnsembler.fromJson(json);
      default:
        return NopEnsembler.fromJson(json);
    }
  }

  toJSON() {
    switch (this.type) {
      case "docker":
        return { type: this.type, docker_config: this.docker_config };
      case "standard":
        if (this.standard_config.experiment_mappings?.length === 0) {
          delete this.standard_config.experiment_mappings;
        }
        if (this.standard_config.route_name_path === "") {
          delete this.standard_config.route_name_path;
        }
        return { type: this.type, standard_config: this.standard_config };
      case "pyfunc":
        return { type: this.type, pyfunc_config: this.pyfunc_config };
      default:
        return { ...this, nop_config: this.nop_config };
    }
  }
}
