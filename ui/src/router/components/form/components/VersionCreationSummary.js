import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const VersionCreationSummary = ({ router }) => {
  return (
    <Fragment>
      <p>
        You're about to create a new version for the Turing router{" "}
        <b>{router.name}</b> in the <b>{router.environment_name}</b> environment
        without deploying it.
      </p>

      <EuiSpacer size="s" />
    </Fragment>
  );
};
