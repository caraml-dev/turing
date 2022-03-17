import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const VersionCreationSummary = ({ router }) => {
  return (
    <Fragment>
      <p>
        You're about to <b>save your changes</b> as a new version for the Turing
        router <b>{router.name}</b>.
      </p>

      <p>
        The new version can be deployed at any time from the router history
        page.
      </p>

      <EuiSpacer size="s" />
    </Fragment>
  );
};
