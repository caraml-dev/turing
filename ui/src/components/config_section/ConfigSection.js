import React, { Fragment } from "react";
import { ConfigSectionTitle } from ".";

export const ConfigSection = ({ title, iconType, children }) => (
  <Fragment>
    <ConfigSectionTitle title={title} iconType={iconType} />
    {children}
  </Fragment>
);
