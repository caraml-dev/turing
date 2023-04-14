import React, { Fragment, useContext } from "react";
import { EuiPageTemplate } from "@elastic/eui";
import { ProjectsContext } from "@caraml-dev/ui-lib";
import { Navigate } from "react-router-dom";
import { useConfig } from "./config";

const Home = () => {
  const { appConfig } = useConfig();
  const { currentProject } = useContext(ProjectsContext);

  return !currentProject ? (
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
  ) : (
    <Navigate to={`projects/${currentProject.id}`} replace={true} />
  );
};

export default Home;
