import React from "react";
import { EuiBadge, EuiBadgeGroup, EuiLink } from "@elastic/eui";
import useCollapse from "react-collapsed";

import "./CollapsibleBadgeGroup.scss";

export const CollapsibleBadgeGroup = ({
  items,
  onClick,
  color = "hollow",
  minItemsCount,
  ...props
}) => {
  const { getToggleProps, isExpanded } = useCollapse();

  const children = () => {
    let children = [];

    if (!!items) {
      children = items.map(
        (item, idx) =>
          (isExpanded || idx < minItemsCount) && (
            <EuiBadge
              key={idx}
              color={color}
              onClick={!!onClick ? () => onClick(item) : undefined}>
              {item}
            </EuiBadge>
          )
      );

      if (!isExpanded && items.length > minItemsCount) {
        children.push(<EuiLink {...getToggleProps()}>Show All</EuiLink>);
      }
    }

    return children;
  };

  return (
    <EuiBadgeGroup className={`euiBadgeGroup---collapsible ${props.className}`}>
      {children()}
    </EuiBadgeGroup>
  );
};
