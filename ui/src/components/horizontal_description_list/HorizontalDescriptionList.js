import { EuiDescriptionList, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import React from "react";

export const HorizontalDescriptionList = ({
  listItems,
  titleProps,
  descriptionProps,
  ...props
}) => (
  <EuiFlexGroup
    direction="row"
    justifyContent="spaceBetween"
    gutterSize={props.gutterSize || "m"}>
    {listItems.map((item, idx) => (
      <EuiFlexItem {...item.flexProps} key={idx}>
        <EuiDescriptionList
          titleProps={titleProps}
          descriptionProps={descriptionProps}
          listItems={[item]}
        />
      </EuiFlexItem>
    ))}
  </EuiFlexGroup>
);
