import React, { Fragment } from "react";
import { EuiPageTemplate } from "@elastic/eui";
import { useConfig } from "./config";

const Home = () => {
  const { appConfig } = useConfig();

  return (
    <EuiPageTemplate>
      <EuiPageTemplate.EmptyPrompt
        iconType={appConfig.appIcon}
        title={<h2>Turing: ML Experiments</h2>}
        body={
          <Fragment>
            <p>To start off, please select a project from the dropdown.</p>
          </Fragment>
        }
      />
    </EuiPageTemplate>
  );
};

export default Home;
