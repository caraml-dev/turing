const moment = require("moment");

export class LogEntry {
  static fromJson(json) {
    return Object.assign(new LogEntry(), json);
  }

  constructor() {
    this.timestamp = "";
    this.pod_name = "";
    this.text_payload = "";
    this.json_payload = null;
  }

  localTimestamp() {
    return moment
      .utc(this.timestamp)
      .local()
      .format("YYYY-MM-DDTHH:mm:ss.SSSZ");
  }

  toString() {
    return `${this.localTimestamp()} - ${
      this.json_payload ? JSON.stringify(this.json_payload) : this.text_payload
    }`;
  }
}
