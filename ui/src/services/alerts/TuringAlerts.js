const objectAssignDeep = require(`object-assign-deep`);

export class TuringAlerts {
  constructor() {
    this.team = undefined;
    this.alerts = {
      throughput: undefined,
      latency95p: undefined,
      error_rate: undefined,
      cpu_util: undefined,
      memory_util: undefined,
    };
  }

  static fromJson(json) {
    return objectAssignDeep(new TuringAlerts(), json);
  }
}
