import React from "react";
import { ConfigMultiSectionPanel } from "../config_multi_section_panel/ConfigMultiSectionPanel";

export const ConfigSectionPanel = (props) => (
  <ConfigMultiSectionPanel
    items={[{ title: props.title, children: props.children }]}
    className={props.className}
  />
);
