const EnumValue = (name, props) =>
  Object.freeze({
    ...props,
    toJSON: () => name,
    toString: () => name
  });

export const Status = Object.freeze({
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
  }),
  fromValue: name =>
    [Status.DEPLOYED, Status.UNDEPLOYED, Status.FAILED, Status.PENDING].find(
      s => name === s.toString()
    )
});
