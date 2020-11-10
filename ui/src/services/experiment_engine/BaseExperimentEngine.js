import { NopExperimentEngine } from "./index";
import { ExperimentEngine } from "./index";

export class BaseExperimentEngine {
  constructor(type) {
    this.type = type;
    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(json = {}) {
    if (!json.type || json.type === "nop") {
      return new NopExperimentEngine();
    }
    return ExperimentEngine.fromJson(json.type, json);
  }

  toJSON() {
    if (this.type === "nop") {
      return { type: this.type };
    }
    return { type: this.type, config: this.config };
  }
}
