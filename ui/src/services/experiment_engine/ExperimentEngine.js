import { BaseExperimentEngine } from "./index";
const objectAssignDeep = require(`object-assign-deep`);

export class ExperimentEngine extends BaseExperimentEngine {
  constructor(type) {
    super(type);
    this.config = {
      client: {},
      experiments: [],
      variables: { config: [] }
    };

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(type, json = {}) {
    const engine = new ExperimentEngine(type);
    engine.config = objectAssignDeep({}, json.config);
    if (engine.config.client && engine.config.client.passkey) {
      engine.config.client.passkey_set = true;
      engine.config.client.passkey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx";
    }

    return engine;
  }

  toJSON() {
    const json = objectAssignDeep({}, this);

    if (json.config.client.passkey_set) {
      delete json.config.client.passkey;
    }
    delete json.config.client.passkey_set;

    return json;
  }
}
