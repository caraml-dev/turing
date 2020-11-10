import { Ensembler } from "./index";

const objectAssignDeep = require(`object-assign-deep`);

export class StandardEnsembler extends Ensembler {
  constructor() {
    super("standard");
    this.standard_config = StandardEnsembler.newConfig();

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json = {}) {
    const ensembler = new StandardEnsembler();
    ensembler.standard_config = objectAssignDeep({}, json.standard_config);

    return ensembler;
  }

  static newConfig() {
    return {
      experiment_mappings: []
    };
  }
}

export const newMapping = (experimentName, treatment) => ({
  experiment: experimentName,
  treatment: treatment,
  route: ""
});
