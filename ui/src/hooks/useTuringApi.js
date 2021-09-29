import { AuthContext, useApi } from "@gojek/mlp-ui";
import { useContext } from "react";
import { useConfig } from "../config";

export const useTuringApi = (
  endpoint,
  options,
  result,
  callImmediately = true
) => {
  const { apiConfig } = useConfig();
  const authCtx = useContext(AuthContext);

  return useApi(
    endpoint,
    {
      baseApiUrl: apiConfig.turingApiUrl,
      timeout: apiConfig.apiTimeout,
      useMockData: apiConfig.useMockData,
      ...options,
    },
    authCtx,
    result,
    callImmediately
  );
};
