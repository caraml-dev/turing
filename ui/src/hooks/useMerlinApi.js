import { AuthContext, useApi } from "@gojek/mlp-ui";
import { useContext } from "react";
import { apiConfig } from "../config";

export const useMerlinApi = (
  endpoint,
  options,
  result,
  callImmediately = true
) => {
  const authCtx = useContext(AuthContext);

  return useApi(
    endpoint,
    {
      baseApiUrl: apiConfig.merlinApiUrl,
      timeout: apiConfig.apiTimeout,
      useMockData: apiConfig.useMockData,
      ...options,
    },
    authCtx,
    result,
    callImmediately
  );
};
