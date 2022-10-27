import React, { useContext, useMemo } from "react";
import { AccordionForm, FormContext } from "@gojek/mlp-ui";
import { RouterStep } from "./steps/RouterStep";
import schema from "./validation/schema";
import { ExperimentStep } from "./steps/ExperimentStep";
import { EnricherStep } from "./steps/EnricherStep";
import { EnsemblerStep } from "./steps/EnsemblerStep";
import { OutcomeStep } from "./steps/OutcomeStep";
import { ConfigSectionTitle } from "../../../components/config_section";
import ExperimentEngineContext from "../../../providers/experiments/context";
import { useConfig } from "../../../config";

export const UpdateRouterForm = ({ projectId, onCancel, onNext }) => {
  const {
    appConfig: {
      scaling: { maxAllowedReplica },
    },
  } = useConfig();

  const validationSchema = useMemo(
    () => schema(maxAllowedReplica),
    [maxAllowedReplica]
  );

  const { experimentEngineOptions, getEngineProperties } = useContext(
    ExperimentEngineContext
  );

  const {
    data: {
      config: { protocol },
    },
  } = useContext(FormContext);

  const sections = [
    {
      title: "Router",
      iconType: "bolt",
      children: <RouterStep projectId={projectId} />,
      validationSchema: validationSchema[0],
      validationContext: { protocol },
    },
    {
      title: "Experiments",
      iconType: "beaker",
      children: <ExperimentStep projectId={projectId} />,
      validationSchema: validationSchema[1],
      validationContext: { experimentEngineOptions, getEngineProperties },
    },
    {
      title: "Enricher",
      iconType: "package",
      children: <EnricherStep projectId={projectId} />,
      validationSchema: validationSchema[2],
    },
    {
      title: "Ensembler",
      iconType: "aggregate",
      children: <EnsemblerStep projectId={projectId} />,
      validationSchema: validationSchema[3],
    },
    {
      title: "Outcome Tracking",
      iconType: "visTagCloud",
      children: <OutcomeStep projectId={projectId} />,
      validationSchema: validationSchema[4],
    },
  ];

  return (
    <AccordionForm
      name="Edit Router"
      sections={sections}
      onCancel={onCancel}
      onSubmit={onNext}
      submitLabel="Next"
      renderTitle={(title, iconType) => (
        <ConfigSectionTitle title={title} iconType={iconType} />
      )}
    />
  );
};
