import { useEffect, useMemo, useState } from "react";
import mitt from "mitt";
import { useTuringPollingApi } from "./useTuringPollingApi";

export const useTuringPollingApiEmitter = (
  apiEndpoint,
  options,
  pollInterval
) => {
  const [isPolling, setIsPolling] = useState(false);

  const [{ data, error }, startPolling, stopPolling] = useTuringPollingApi(
    apiEndpoint,
    options,
    undefined,
    pollInterval
  );

  const emitter = useMemo(() => {
    const emitter = new mitt();

    emitter.on("start", () => {
      setIsPolling(true);
    });

    emitter.on("abort", () => {
      setIsPolling(false);
    });

    return emitter;
  }, [setIsPolling]);

  useEffect(() => {
    if (isPolling) {
      startPolling();
      return stopPolling;
    }
  }, [isPolling, startPolling, stopPolling]);

  useEffect(() => {
    if (!!data) {
      emitter.emit("data", data);
    }
  }, [emitter, data]);

  useEffect(() => {
    if (!!error) {
      emitter.emit("error", error);
    }
  }, [emitter, error]);

  return { emitter };
};
