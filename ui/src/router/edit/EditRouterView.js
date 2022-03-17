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
  const { data: routerConfig } = useContext(FormContext);
  const [showDiffView, toggleDiffView] = useToggle();

  const [updateRouterResponse, updateRouterForm] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  const [createRouterVersionResponse, createRouterVersionForm] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}/versions`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (updateRouterResponse.isLoaded && !updateRouterResponse.error) {
      addToast({
        id: "submit-success-update-router",
        title: "Router configuration is updated!",
        color: "success",
        iconType: "check",
      });

      props.navigate("../", { state: { refresh: true } });
    }
  }, [updateRouterResponse, props]);

  useEffect(() => {
    if (
      createRouterVersionResponse.isLoaded &&
      !createRouterVersionResponse.error
    ) {
      addToast({
        id: "submit-success-create-router-version",
        title: "Router configuration is created!",
        color: "success",
        iconType: "check",
      });

      props.navigate("../", { state: { refresh: true } });
    }
  }, [createRouterVersionResponse, props]);

  const [withDeployment, setWithDeployment] = useState(null);

  const onSubmit = () => {
    if (withDeployment === true) {
      return updateRouterForm({
        body: JSON.stringify(routerConfig),
      });
    } else if (withDeployment === false) {
      return createRouterVersionForm({
        body: JSON.stringify(routerConfig),
      });
    }
  };

  return (
    <ConfirmationModal
      title="Update Turing Router"
      content={
        withDeployment ? (
          <DeploymentSummary router={routerConfig} />
        ) : (
          <VersionCreationSummary router={routerConfig} />
        )
      }
      isLoading={
        updateRouterResponse.isLoading || createRouterVersionResponse.isLoading
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
            updatedRouter={routerConfig}
            onPrevious={toggleDiffView}
            onSubmit={onSubmit}
            isSubmitting={
              updateRouterResponse.isLoading ||
              createRouterVersionResponse.isLoading
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
