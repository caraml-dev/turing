import React, { useContext, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { TuringRouter } from "../../services/router/TuringRouter";
import { FormContext, FormContextProvider } from "@gojek/mlp-ui";
import { addToast, replaceBreadcrumbs, useToggle } from "@gojek/mlp-ui";
import { UpdateRouterForm } from "../components/form/UpdateRouterForm";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";
import { VersionComparisonView } from "./components/VersionComparisonView";
import { useTuringApi } from "../../hooks/useTuringApi";
import { RouterUpdateSummary } from "../components/form/components/RouterUpdateSummary";
import { ConfirmationModal } from "../../components/confirmation_modal/ConfirmationModal";
import { VersionCreationSummary } from "../components/form/components/VersionCreationSummary";

const EditRouterView = ({ projectId, currentRouter }) => {
  const navigate = useNavigate();
  const { data: routerConfig } = useContext(FormContext);
  const [showDiffView, toggleDiffView] = useToggle();

  const [updateRouterResponse, submitUpdateRouter] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  const [createRouterVersionResponse, submitCreateRouterVersion] = useTuringApi(
    `/projects/${projectId}/routers/${currentRouter.id}/versions`,
    {
      method: "POST",
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

      navigate("../", { state: { refresh: true } });
    }
  }, [updateRouterResponse, navigate]);

  useEffect(() => {
    if (
      createRouterVersionResponse.isLoaded &&
      !createRouterVersionResponse.error
    ) {
      addToast({
        id: "submit-success-create-router-version",
        title: `Router version ${createRouterVersionResponse.data.version} is saved (but not deployed)!`,
        color: "success",
        iconType: "check",
      });

      navigate("../", { state: { refresh: true } });
    }
  }, [createRouterVersionResponse, navigate]);

  const [withDeployment, setWithDeployment] = useState(null);

  const onSubmit = () => {
    if (withDeployment === true) {
      return submitUpdateRouter({
        body: JSON.stringify(routerConfig),
      });
    } else if (withDeployment === false) {
      return submitCreateRouterVersion({
        body: JSON.stringify(routerConfig.toJSON().config),
      });
    }
  };

  return (
    <ConfirmationModal
      title={
        withDeployment
          ? "Deploy Turing Router Version"
          : "Save Turing Router Version"
      }
      content={
        withDeployment ? (
          <RouterUpdateSummary router={routerConfig} />
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
            onCancel={() => navigate("../")}
            onNext={toggleDiffView}
          />
        ) : (
          <VersionComparisonView
            currentRouter={currentRouter}
            updatedRouter={routerConfig}
            onPrevious={toggleDiffView}
            isSubmitting={
              updateRouterResponse.isLoading ||
              createRouterVersionResponse.isLoading
            }
            onDeploy={() => {
              setWithDeployment(true);
              return onSubmit();
            }}
            onSave={() => {
              setWithDeployment(false);
              return onSubmit();
            }}
          />
        );
      }}
    </ConfirmationModal>
  );
};

const EditRouterViewWrapper = ({ router }) => {
  const { projectId } = useParams();
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
        />
      </ExperimentEngineContextProvider>
    </FormContextProvider>
  );
};

export { EditRouterViewWrapper as EditRouterView };
