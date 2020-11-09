import { BaseExperimentEngine } from "./index";

export class NopExperimentEngine extends BaseExperimentEngine {
  constructor() {
    super("nop");
  }
}
