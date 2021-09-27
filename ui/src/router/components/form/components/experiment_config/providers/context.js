import React from "react";

const ExperimentContext = React.createContext({
  clients: [],
  experiments: [],
  variables: {},
});

export default ExperimentContext;
