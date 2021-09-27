import { Enum, EnumValue } from "../enum/Enum";

export const JobStatus = Enum({
  PENDING: EnumValue("pending", {
    label: "Pending",
    color: "#F5A700",
    iconType: "clock",
  }),
  BUILDING: EnumValue("building", {
    label: "Pending",
    color: "#F5A700",
    iconType: "clock",
  }),
  RUNNING: EnumValue("running", {
    label: "Running",
    color: "#017D73",
    iconType: "check",
  }),
  TERMINATING: EnumValue("terminating", {
    label: "Terminating",
    color: "#BD271E",
    iconType: "clock",
  }),
  TERMINATED: EnumValue("terminated", {
    label: "Terminated",
    color: "#6A717D",
  }),
  COMPLETED: EnumValue("completed", {
    label: "Completed",
    color: "#6A717D",
    iconType: "check",
  }),
  FAILED: EnumValue("failed", {
    label: "Failed",
    color: "#BD271E",
    iconType: "cross",
  }),
  FAILED_SUBMISSION: EnumValue("failed_submission", {
    label: "Failed",
    color: "#BD271E",
    iconType: "cross",
  }),
  FAILED_BUILDING: EnumValue("failed_building", {
    label: "Failed",
    color: "#BD271E",
    iconType: "cross",
  }),
});
