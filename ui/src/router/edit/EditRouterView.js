import React, { useContext, useEffect } from "react";
import { TuringRouter } from "../../services/router/TuringRouter";
import {
  FormContext,
  FormContextProvider,
} from "../../components/form/context";
import { addToast, replaceBreadcrumbs, useToggle } from "@gojek/mlp-ui";
import { UpdateRouterForm } from "../components/form/UpdateRouterForm";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";
import { VersionComparisonView } from "./components/VersionComparisonView";
import { useTuringApi } from "../../hooks/useTuringApi";
import { DeploymentSummary } from "../components/form/components/DeploymentSummary";
import { ConfirmationModal } from "../../components/confirmation_modal/ConfirmationModal";

const EditRouterView = ({ projectId, currentRouter, ...props }) => {
  const { data: updatedRouter } = useContext(FormContext);
  const [showDiffView, toggleDiffView] = useToggle();

  const [submissionResponse, submitForm] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-create",
        title: "Router configuration is updated!",
        color: "success",
        iconType: "check",
      });

      props.navigate("../", { state: { refresh: true } });
    }
  }, [submissionResponse, props]);

  const onSubmit = () => submitForm({ body: JSON.stringify(updatedRouter) });

  return (
    <ConfirmationModal
      title="Update Turing Router"
      content={<DeploymentSummary router={updatedRouter} />}
      isLoading={submissionResponse.isLoading}
      onConfirm={onSubmit}
      confirmButtonText="Deploy"
      confirmButtonColor="primary">
      {(onSubmit) =>
        !showDiffView ? (
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
            isSubmitting={submissionResponse.isLoading}
          />
        )
      }
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
