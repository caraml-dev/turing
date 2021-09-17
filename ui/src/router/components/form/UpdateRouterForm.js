import React, { useContext, useEffect } from "react";
import { FormContext } from "../../../components/form/context";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { addToast } from "@gojek/mlp-ui";
import { DeploymentSummary } from "./components/DeploymentSummary";
import { ConfirmationModal } from "../../../components/confirmation_modal/ConfirmationModal";
import { AccordionForm } from "../../../components/accordion_form";
import { RouterStep } from "./steps/RouterStep";
import schema from "./validation/schema";
import { ExperimentStep } from "./steps/ExperimentStep";
import { EnricherStep } from "./steps/EnricherStep";
import { EnsemblerStep } from "./steps/EnsemblerStep";
import { OutcomeStep } from "./steps/OutcomeStep";
import { ConfigSectionTitle } from "../configuration/components/section";
import ExperimentEngineContext from "../../../providers/experiments/context";

export const UpdateRouterForm = ({ projectId, onCancel, onSuccess }) => {
  const { data: router } = useContext(FormContext);

  const [submissionResponse, submitForm] = useTuringApi(
    `/projects/${projectId}/routers/${router.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  const { experimentEngineOptions } = useContext(ExperimentEngineContext);

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-create",
        title: "Router configuration is updated!",
        color: "success",
        iconType: "check",
      });
      onSuccess();
    }
  }, [submissionResponse, onSuccess]);

  const onSubmit = () => submitForm({ body: JSON.stringify(router) });

  const sections = [
    {
      title: "Router",
      iconType: "bolt",
      children: <RouterStep projectId={projectId} />,
      validationSchema: schema[0],
    },
    {
      title: "Experiments",
      iconType: "beaker",
      children: <ExperimentStep projectId={projectId} />,
      validationSchema: schema[1],
      validationContext: { experimentEngineOptions },
    },
    {
      title: "Enricher",
      iconType: "package",
      children: <EnricherStep projectId={projectId} />,
      validationSchema: schema[2],
    },
    {
      title: "Ensembler",
      iconType: "aggregate",
      children: <EnsemblerStep projectId={projectId} />,
      validationSchema: schema[3],
    },
    {
      title: "Outcome Tracking",
      iconType: "visTagCloud",
      children: <OutcomeStep projectId={projectId} />,
      validationSchema: schema[4],
    },
  ];

  return (
    <ConfirmationModal
      title="Update Turing Router"
      content={<DeploymentSummary router={router} />}
      isLoading={submissionResponse.isLoading}
      onConfirm={onSubmit}
      confirmButtonText="Deploy"
      confirmButtonColor="primary">
      {(onSubmit) => (
        <AccordionForm
          name="Edit Router"
          sections={sections}
          onCancel={onCancel}
          onSubmit={onSubmit}
          submitLabel="Deploy"
          renderTitle={(title, iconType) => (
            <ConfigSectionTitle title={title} iconType={iconType} />
          )}
        />
      )}
    </ConfirmationModal>
  );
};
