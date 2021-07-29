import { EnumValue } from "../enum_value/EnumValue";

const values = {
  DEPLOYED: EnumValue("deployed", {
    label: "Deployed",
    color: "#017D73",
    iconType: "check"
  }),
  FAILED: EnumValue("failed", {
    label: "Failed",
    color: "#BD271E",
    iconType: "cross"
  }),
  PENDING: EnumValue("pending", {
    label: "Updating",
    color: "#F5A700",
    iconType: "clock"
  }),
  UNDEPLOYED: EnumValue("undeployed", {
    label: "Not Deployed",
    color: "#6A717D"
  })
};

const allValues = Object.values(values);

export const Status = Object.freeze({
  ...values,
  values: allValues,
  fromValue: name => allValues.find(s => name === s.toString())
});
