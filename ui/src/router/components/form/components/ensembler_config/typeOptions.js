import React, { Fragment } from "react";

const typeOptions = {
  HTTP_JSON: [
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
      description: (
        <Fragment>
          Turing will select the response from one of the routes, based on the
          configured mapping between routes and experiment treatments.
        </Fragment>
      ),
    },
    {
      value: "docker",
      inputDisplay: "Docker",
      description: (
        <Fragment>
          Turing will deploy the specified image as a post-processor and send
          the responses from all routes, together with the treatment
          configuration, for the ensembling the final response.
        </Fragment>
      ),
    },
    {
      value: "pyfunc",
      inputDisplay: "Pyfunc",
      description: (
        <Fragment>
          Turing will build and deploy the selected pyfunc ensembler and send
          the responses from all routes, together with the treatment
          configuration, for ensembling the final response.
        </Fragment>
      ),
    },
    {
      value: "external",
      inputDisplay: "External (Coming Soon)",
      description: (
        <Fragment>
          Turing will send the responses from all routes, along with treatment
          configuration to the external URL, for ensembling the final response.
        </Fragment>
      ),
      disabled: true,
    },
  ],
  UPI_V1: [
    {
      value: "nop",
      inputDisplay: "None",
      description: (
        <Fragment>
          Turing will simply return the response from the route marked as the{" "}
          <strong>Final Response</strong>. Other ensembler types are not yet
          supported for the UPI protocol.
        </Fragment>
      ),
    },
    {
      value: "standard",
      inputDisplay: "Standard",
      description: (
        <Fragment>
          Turing will select the response from one of the routes, based on the
          configured mapping between routes and experiment treatments. Only
          standard ensembler is supported for UPI now.
        </Fragment>
      ),
    },
  ],
};

export const ensemblerTypeOptions = (engineProps, protocol) => {
  const options = typeOptions[protocol];
  if (!engineProps.name) {
    // Standard Ensembler is not available when there is no experiment engine
    return options.filter((o) => o.value !== "standard");
  }
  // Ensembler must be selected when there is an experiment engine
  const ensemblerOptions = options.filter((o) => o.value !== "nop");
  if (
    engineProps?.standard_experiment_manager_config
      ?.experiment_selection_enabled === false
  ) {
    // Standard Ensembler is not available when experiment selection is disabled
    return ensemblerOptions.filter((o) => o.value !== "standard");
  }
  return ensemblerOptions;
};
