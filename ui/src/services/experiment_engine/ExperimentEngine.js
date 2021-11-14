import { BaseExperimentEngine } from "./index";
const objectAssignDeep = require(`object-assign-deep`);

export class ExperimentEngine extends BaseExperimentEngine {
  constructor(type) {
    super(type);
    this.config = {
      client: {},
      experiments: [],
      variables: { config: [] },
    };

    this.toJSON = this.toJSON.bind(this);
  }

  static fromJson(type, json = {}) {
    const engine = new ExperimentEngine(type);
    engine.config = objectAssignDeep({}, json.config);

    if (engine.config.client && engine.config.client?.passkey) {
      if (typeof engine.config.client.passkey_set === "undefined") {
        engine.config.client.passkey_set = true;
        engine.config.client.encrypted_passkey = engine.config.client.passkey;
        engine.config.client.passkey = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx";
      } else if (engine.config.client?.passkey_set === false) {
        delete engine.config.client.encrypted_passkey;
      }
    }

    return engine;
  }

  toJSON() {
    const json = objectAssignDeep({}, this);

    if (json.config?.client?.passkey_set) {
      delete json.config.client.passkey;
    }
    if (json.config?.client?.passkey_set) {
      delete json.config.client.passkey_set;
    }

    return json;
  }
}
