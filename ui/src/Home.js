import React, { Fragment } from "react";
import { EuiEmptyPrompt } from "@elastic/eui";
import { appConfig } from "./config";

const Home = () => (
  <EuiEmptyPrompt
    iconType={appConfig.appIcon}
    title={<h2>Turing: ML Experiments</h2>}
    body={
      <Fragment>
        <p>To start off, please select a project from the dropdown.</p>
      </Fragment>
    }
  />
);

export default Home;
