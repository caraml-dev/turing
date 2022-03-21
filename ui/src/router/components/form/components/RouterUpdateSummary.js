import React, { Fragment } from "react";
import { EuiSpacer } from "@elastic/eui";

export const RouterUpdateSummary = ({ router }) => {
  return (
    <Fragment>
      <p>
        You're about to <b>deploy</b> your changes for the Turing router{" "}
        <b>{router.name}</b>.
      </p>

      <EuiSpacer size="s" />
    </Fragment>
  );
};
