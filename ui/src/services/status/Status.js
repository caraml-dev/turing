import { EnumValue, Enum } from "../enum/Enum";

export const Status = Enum({
  DEPLOYED: EnumValue("deployed", {
    label: "Deployed",
    color: "success",
    iconType: "check",
  }),
  FAILED: EnumValue("failed", {
    label: "Failed",
    color: "danger",
    iconType: "cross",
  }),
  PENDING: EnumValue("pending", {
    label: "Updating",
    color: "warning",
    iconType: "clock",
  }),
  UNDEPLOYED: EnumValue("undeployed", {
    label: "Not Deployed",
    color: "#6A717D",
  }),
});
