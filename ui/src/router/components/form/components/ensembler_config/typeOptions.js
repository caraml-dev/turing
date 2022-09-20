import React, { Fragment } from "react";

const typeOptions = [
  {
    value: "nop",
    inputDisplay: "None",
    description: (
      <Fragment>
        Turing will simply return response from the route marked as the{" "}
        <strong>Final Response</strong>
      </Fragment>
    ),
  },
  {
    value: "standard",
    inputDisplay: "Standard",
    expenginetype: "standard",
    description: (
      <Fragment>
        Turing will select the response from one of the routes, based on the
        configured mapping between routes and experiment treatments
      </Fragment>
    ),
  },
  {
    value: "standard",
    inputDisplay: "Standard",
    expenginetype: "custom",
    description: (
      <Fragment>
        Turing will select the route response corresponding to the route name specified
        in the treatment configuration. The route name path will be used to locate the
        name of the route within the treatment configuration.
      </Fragment>
    ),
  },
  {
    value: "docker",
    inputDisplay: "Docker",
    description: (
      <Fragment>
        Turing will deploy specified image as a post-processor and will send to
        it responses from all routes, together with the treatment configuration,
        for the ensembling
      </Fragment>
    ),
  },
  {
    value: "pyfunc",
    inputDisplay: "Pyfunc",
    description: (
      <Fragment>
        Turing will build and deploy the selected pyfunc ensembler and will send
        to it responses from all routes, together with the treatment
        configuration, for ensembling
      </Fragment>
    ),
  },
  {
    value: "external",
    inputDisplay: "External (Coming Soon)",
    description: (
      <Fragment>
        Turing will send responses from all routes, together with treatment
        configuration, to the external URL for ensembling
      </Fragment>
    ),
    disabled: true,
  },
];

export const ensemblerTypeOptions = (engineProps) => {
  if (!engineProps.name) {
    // Standard Ensembler is not available when there is no experiment engine
    return typeOptions.filter((o) => o.value !== "standard");
  }
  // Ensembler must be selected when there is an experiment engine
  const ensemblerOptions = typeOptions.filter(
    (o) => o.value !== "nop" && (o.expenginetype === engineProps.type || o.expenginetype === undefined)
  );
  if (engineProps?.standard_experiment_manager_config?.experiment_selection_enabled === false) {
    // Standard Ensembler is not available when experiment selection is disabled
    return ensemblerOptions.filter((o) => o.value !== "standard");
  }
  return ensemblerOptions;
};
