import React, { useEffect } from "react";
import { replaceBreadcrumbs, FormContextProvider } from "@gojek/mlp-ui";
import { TuringRouter } from "../../services/router/TuringRouter";
import { CreateRouterForm } from "../components/form/CreateRouterForm";
import { EuiPageTemplate } from "@elastic/eui";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";

export const CreateRouterView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([{ text: `Routers`, href: "." }, { text: "Create" }]);
  }, [projectId]);

  return (
    <EuiPageTemplate>
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"graphApp"}
        pageTitle="Create Router"
      />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={new TuringRouter()}>
          <ExperimentEngineContextProvider>
            <CreateRouterForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={(routerId) => props.navigate(`../${routerId}`)}
            />
          </ExperimentEngineContextProvider>
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};
