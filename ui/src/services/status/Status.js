import { EnumValue, Enum } from "../enum/Enum";

export const Status = Enum({
  DEPLOYED: EnumValue("deployed", {
    label: "Deployed",
    color: "#017D73",
    iconType: "check",
  }),
  FAILED: EnumValue("failed", {
    label: "Failed",
    color: "#BD271E",
    iconType: "cross",
  }),
  PENDING: EnumValue("pending", {
    label: "Updating",
    color: "#F5A700",
    iconType: "clock",
  }),
  UNDEPLOYED: EnumValue("undeployed", {
    label: "Not Deployed",
    color: "#6A717D",
  }),
});
