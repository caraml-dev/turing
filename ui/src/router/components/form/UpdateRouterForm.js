import React, { useContext } from "react";
import { AccordionForm } from "../../../components/accordion_form";
import { RouterStep } from "./steps/RouterStep";
import schema from "./validation/schema";
import { ExperimentStep } from "./steps/ExperimentStep";
import { EnricherStep } from "./steps/EnricherStep";
import { EnsemblerStep } from "./steps/EnsemblerStep";
import { OutcomeStep } from "./steps/OutcomeStep";
import { ConfigSectionTitle } from "../../../components/config_section";
import ExperimentEngineContext from "../../../providers/experiments/context";

export const UpdateRouterForm = ({ projectId, onCancel, onNext }) => {
  const { experimentEngineOptions } = useContext(ExperimentEngineContext);

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
      children: <ExperimentStep />,
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
