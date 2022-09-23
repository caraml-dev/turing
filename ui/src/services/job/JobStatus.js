import { Enum, EnumValue } from "../enum/Enum";

export const JobStatus = Enum({
  PENDING: EnumValue("pending", {
    label: "Pending",
    color: "warning",
    iconType: "clock",
  }),
  BUILDING: EnumValue("building", {
    label: "Pending",
    color: "warning",
    iconType: "clock",
  }),
  RUNNING: EnumValue("running", {
    label: "Running",
    color: "success",
    iconType: "check",
  }),
  TERMINATING: EnumValue("terminating", {
    label: "Terminating",
    color: "danger",
    iconType: "clock",
  }),
  TERMINATED: EnumValue("terminated", {
    label: "Terminated",
    color: "default",
  }),
  COMPLETED: EnumValue("completed", {
    label: "Completed",
    color: "default",
    iconType: "check",
  }),
  FAILED: EnumValue("failed", {
    label: "Failed",
    color: "danger",
    iconType: "cross",
  }),
  FAILED_SUBMISSION: EnumValue("failed_submission", {
    label: "Failed",
    color: "danger",
    iconType: "cross",
  }),
  FAILED_BUILDING: EnumValue("failed_building", {
    label: "Failed",
    color: "danger",
    iconType: "cross",
  }),
});
