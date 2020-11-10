import { Ensembler } from "./Ensembler";

export class NopEnsembler extends Ensembler {
  constructor() {
    super("nop");
  }
}
