import { useTuringApi } from "./useTuringApi";
import { useCallback, useRef } from "react";

export const usePollingTuringApi = (apiUrl, options, result, pollInterval) => {
  const [response, fetchOnce] = useTuringApi(apiUrl, options, result, false);
  const _stopPolling = useRef(() => {});

  const startPolling = useCallback(
    options => {
      const promise = fetchOnce(options);
      const timeout = setTimeout(() => startPolling(options), pollInterval);

      _stopPolling.current = () => {
        promise.cancel();
        clearTimeout(timeout);
      };
    },
    [pollInterval, fetchOnce]
  );

  const stopPolling = useCallback(() => {
    _stopPolling.current();
  }, []);

  return [response, startPolling, stopPolling, fetchOnce];
};
