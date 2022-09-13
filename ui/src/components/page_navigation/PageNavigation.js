import React, { useState } from "react";
import {
  EuiContextMenuItem,
  EuiContextMenuPanel,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiPopover,
  EuiTab,
  EuiTabs,
} from "@elastic/eui";

import "./PageNavigation.scss";

export const PageNavigation = ({
  tabs,
  actions,
  selectedTab = "",
  ...props
}) => (
  <EuiFlexGroup direction="row" gutterSize="none">
    <EuiFlexItem grow={true}>
      <EuiTabs>
        {tabs.map((tab, index) => (
          <EuiTab
            {...(tab.href
              ? { href: tab.href, target: "_blank" }
              : { onClick: () => props.navigate(`./${tab.id}`) })}
            isSelected={selectedTab.startsWith(tab.id)}
            disabled={tab.disabled}
            key={index}>
            {tab.name}
          </EuiTab>
        ))}
      </EuiTabs>
    </EuiFlexItem>
    {actions && actions.length && (
      <EuiFlexItem grow={false}>
        <MoreActionsButton actions={actions} />
      </EuiFlexItem>
    )}
  </EuiFlexGroup>
);

const MoreActionsButton = ({ actions }) => {
  const [isPopoverOpen, setPopover] = useState(false);
  const togglePopover = () => setPopover((isPopoverOpen) => !isPopoverOpen);

  const items = actions
    .filter((item) => !item.hidden)
    .map((item, idx) => (
      <EuiContextMenuItem
        key={idx}
        icon={item.icon}
        onClick={() => {
          togglePopover();
          item.onClick();
        }}
        disabled={item.disabled}
        className={item.color ? `euiTextColor--${item.color}` : ""}>
        {item.name}
      </EuiContextMenuItem>
    ));

  const button = (
    <EuiTabs>
      <EuiTab onClick={togglePopover}>
        <span>
          More Actions&nbsp;
          <EuiIcon
            type="arrowDown"
            size="m"
            style={{ verticalAlign: "text-bottom" }}
          />
        </span>
      </EuiTab>
    </EuiTabs>
  );

  return (
    <EuiPopover
      button={button}
      isOpen={isPopoverOpen}
      closePopover={togglePopover}
      panelPaddingSize="none"
      anchorPosition="downRight">
      <EuiContextMenuPanel
        className="euiContextPanel--moreActions"
        items={items}
      />
    </EuiPopover>
  );
};
