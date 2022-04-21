import { Ensembler } from "./Ensembler";

const objectAssignDeep = require(`object-assign-deep`);

export class NopEnsembler extends Ensembler {
  constructor() {
    super("nop");

    this.nop_config = NopEnsembler.newConfig();
    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json = {}) {
    const ensembler = new NopEnsembler();
    ensembler.nop_config = objectAssignDeep({}, json.nop_config);
    return ensembler;
  }

  static newConfig() {
    return {
      final_response_route_id: "",
    };
  }
}
