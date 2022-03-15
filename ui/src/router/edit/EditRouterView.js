import React, { useContext, useEffect, useState } from "react";
import { TuringRouter } from "../../services/router/TuringRouter";
import { FormContext, FormContextProvider } from "@gojek/mlp-ui";
import { addToast, replaceBreadcrumbs, useToggle } from "@gojek/mlp-ui";
import { UpdateRouterForm } from "../components/form/UpdateRouterForm";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";
import { VersionComparisonView } from "./components/VersionComparisonView";
import { useTuringApi } from "../../hooks/useTuringApi";
import { DeploymentSummary } from "../components/form/components/DeploymentSummary";
import { ConfirmationModal } from "../../components/confirmation_modal/ConfirmationModal";
import { VersionCreationSummary } from "../components/form/components/VersionCreationSummary";

const EditRouterView = ({ projectId, currentRouter, ...props }) => {
  const { data: updatedRouter } = useContext(FormContext);
  const [showDiffView, toggleDiffView] = useToggle();

  const [
    submissionCreateVersionWithDeploymentResponse,
    submitCreateVersionWithDeploymentForm,
  ] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  const [
    submissionCreateVersionWithoutDeploymentResponse,
    submitCreateVersionWithoutDeploymentForm,
  ] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}/versions`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (
      submissionCreateVersionWithDeploymentResponse.isLoaded &&
      !submissionCreateVersionWithDeploymentResponse.error
    ) {
      addToast({
        id: "submit-success-create-version-with-deployment",
        title: "Router configuration is created and sent for deployment!",
        color: "success",
        iconType: "check",
      });

      props.navigate("../", { state: { refresh: true } });
    }
  }, [submissionCreateVersionWithDeploymentResponse, props]);

  useEffect(() => {
    if (
      submissionCreateVersionWithoutDeploymentResponse.isLoaded &&
      !submissionCreateVersionWithoutDeploymentResponse.error
    ) {
      addToast({
        id: "submit-success-create-version-without-deployment",
        title: "Router configuration is created (but not deployed)!",
        color: "success",
        iconType: "check",
      });

      props.navigate("../", { state: { refresh: true } });
    }
  }, [submissionCreateVersionWithoutDeploymentResponse, props]);

  const [withDeployment, setWithDeployment] = useState(null);

  const onSubmit = () => {
    if (withDeployment === true) {
      return submitCreateVersionWithDeploymentForm({
        body: JSON.stringify(updatedRouter),
      });
    } else if (withDeployment === false) {
      return submitCreateVersionWithoutDeploymentForm({
        body: JSON.stringify(updatedRouter),
      });
    }
  };

  return (
    <ConfirmationModal
      title="Update Turing Router"
      content={
        withDeployment ? (
          <DeploymentSummary router={updatedRouter} />
        ) : (
          <VersionCreationSummary router={updatedRouter} />
        )
      }
      isLoading={
        submissionCreateVersionWithDeploymentResponse.isLoading ||
        submissionCreateVersionWithoutDeploymentResponse.isLoading
      }
      onConfirm={onSubmit}
      confirmButtonText={withDeployment ? "Deploy" : "Save"}
      confirmButtonColor={"primary"}>
      {(onSubmit) => {
        return !showDiffView ? (
          <UpdateRouterForm
            projectId={projectId}
            onCancel={() => props.navigate("../")}
            onNext={toggleDiffView}
            {...props}
          />
        ) : (
          <VersionComparisonView
            currentRouter={currentRouter}
            updatedRouter={updatedRouter}
            onPrevious={toggleDiffView}
            onSubmit={onSubmit}
            isSubmitting={
              submissionCreateVersionWithDeploymentResponse.isLoading ||
              submissionCreateVersionWithoutDeploymentResponse.isLoading
            }
            setWithDeployment={setWithDeployment}
          />
        );
      }}
    </ConfirmationModal>
  );
};

const EditRouterViewWrapper = ({ projectId, router, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: `../`,
      },
      {
        text: router.name,
        href: `./`,
      },
      {
        text: "Update",
      },
    ]);
  }, [router]);

  return (
    <FormContextProvider data={TuringRouter.fromJson(router)}>
      <ExperimentEngineContextProvider>
        <EditRouterView
          projectId={projectId}
          currentRouter={router}
          {...props}
        />
      </ExperimentEngineContextProvider>
    </FormContextProvider>
  );
};

export { EditRouterViewWrapper as EditRouterView };
