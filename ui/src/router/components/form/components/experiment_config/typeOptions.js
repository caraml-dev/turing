import React, { Fragment } from "react";

const defaultExperimentEngineOptions = [
  {
    value: "nop",
    inputDisplay: "None",
    description: (
      <Fragment>
        Turing will send the original request to all of the configured routes,
        but response from Turing will depend on the Ensembler configuration.
      </Fragment>
    )
  }
];

export const getExperimentEngineOptions = engines => {
  return [
    ...defaultExperimentEngineOptions,
    ...engines.map(item => ({
      value: item.name.toLowerCase(),
      inputDisplay: item.name
    }))
  ];
};
