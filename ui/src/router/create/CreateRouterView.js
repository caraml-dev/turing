import React, { useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { FormContextProvider } from "../../components/form/context";
import { TuringRouter } from "../../services/router/TuringRouter";
import { CreateRouterForm } from "../components/form/CreateRouterForm";
import {
  EuiPage,
  EuiPageBody,
  EuiPageContentBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
} from "@elastic/eui";
import { PageTitle } from "../../components/page/PageTitle";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";

export const CreateRouterView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([{ text: `Routers`, href: "." }, { text: "Create" }]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Create Router" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <FormContextProvider data={new TuringRouter()}>
            <ExperimentEngineContextProvider>
              <CreateRouterForm
                projectId={projectId}
                onCancel={() => window.history.back()}
                onSuccess={(routerId) => props.navigate(`../${routerId}`)}
              />
            </ExperimentEngineContextProvider>
          </FormContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};
