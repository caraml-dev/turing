import React, { useContext, useEffect, useMemo, useState } from "react";
import EndpointsContext from "./context";
import EnvironmentsContext from "../environments/context";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const getEnvironmentRegion = (envName, environments) => {
  const env = environments.find(e => e.name === envName);
  return env ? env.region : undefined;
};

export const MerlinEndpointsProvider = ({
  projectId,
  environmentName,
  ...props
}) => {
  const environments = useContext(EnvironmentsContext);

  const region = useMemo(
    () => getEnvironmentRegion(environmentName, environments),
    [environmentName, environments]
  );

  const [{ data }, fetchMerlinEndpoints] = useMerlinApi(
    `/projects/${projectId}/model_endpoints?region=${region}`,
    {},
    [],
    false
  );

  const [endpoints, setEndpoints] = useState([]);

  useEffect(() => {
    region && fetchMerlinEndpoints();
  }, [region, fetchMerlinEndpoints]);

  useEffect(() => {
    setEndpoints(data);
  }, [data, setEndpoints]);

  return (
    <EndpointsContext.Provider value={endpoints}>
      {props.children}
    </EndpointsContext.Provider>
  );
};
