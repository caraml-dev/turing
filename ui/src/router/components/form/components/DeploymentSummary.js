import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const DeploymentSummary = ({ router }) => {
  return (
    <Fragment>
      <p>
        You're about to deploy a new Turing router <b>{router.name}</b> into{" "}
        <b>{router.environment_name}</b> environment.
      </p>

      <EuiSpacer size="s" />
    </Fragment>
  );
};
