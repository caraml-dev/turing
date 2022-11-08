import React, { useContext, useEffect, useMemo, useState } from "react";
import EndpointsContext from "./context";
import EnvironmentsContext from "../environments/context";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const getEnvironmentRegion = (envName, environments) => {
  const env = environments.find((e) => e.name === envName);
  return env ? env.region : undefined;
};

export const ANNOTATIONS_MERLIN_MODEL_ID = "merlin.gojek.com/model-id";

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
    const urlWithPortRegex = /.*(:\d+)$/;
    const formatMerlinUrl = (url, protocol) => {
      // Append port 80 as default if Merlin doesn't return url with port
      if (protocol === "UPI_V1") {
        if (url.match(urlWithPortRegex)) {
          return url;
        }
        return `${url}:80`;
      }
      return `http://${url}/v1/predict`;
    };

    setEndpoints([
      {
        label: "Merlin Model Endpoints",
        options: data
          // To only show currently deployed endpoints
          .filter((modelEndpoint) => modelEndpoint.status === "serving")
          .map((modelEndpoint) => ({
            icon: "machineLearningApp",
            label: formatMerlinUrl(modelEndpoint.url, modelEndpoint.protocol),
            annotations: {
              [ANNOTATIONS_MERLIN_MODEL_ID]: `${modelEndpoint.model.id}`,
            },
          })),
      },
    ]);
  }, [data, setEndpoints]);

  return (
    <EndpointsContext.Provider value={endpoints}>
      {props.children}
    </EndpointsContext.Provider>
  );
};
