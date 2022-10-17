import React, { useEffect } from "react";
import { replaceBreadcrumbs, FormContextProvider } from "@gojek/mlp-ui";
import { useNavigate, useParams } from "react-router-dom";
import { TuringRouter } from "../../services/router/TuringRouter";
import { CreateRouterForm } from "../components/form/CreateRouterForm";
import { EuiPageTemplate, EuiSpacer } from "@elastic/eui";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";
import { useConfig } from "../../config";

export const CreateRouterView = () => {
  const { projectId } = useParams();
  const navigate = useNavigate();
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  useEffect(() => {
    replaceBreadcrumbs([{ text: `Routers`, href: "." }, { text: "Create" }]);
  }, [projectId]);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"graphApp"}
        pageTitle="Create Router"
      />

      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={new TuringRouter()}>
          <ExperimentEngineContextProvider>
            <CreateRouterForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={(routerId) => navigate(`../${routerId}`)}
            />
          </ExperimentEngineContextProvider>
        </FormContextProvider>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};
