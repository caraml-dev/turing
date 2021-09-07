import { useCallback, useEffect, useState } from "react";
import { usePollingTuringApi } from "../../hooks/usePollingTuringApi";
import useEventEmitter from "./useEventEmitter";

const POLLING_INTERVAL = 7000;
const BATCH_SIZE = 500;

const useEventEmitterLogsApi = (apiEndpoint, params, processLogs) => {
  const { emitter, isActive } = useEventEmitter();

  const [query, setQuery] = useState();

  useEffect(() => {
    setQuery({
      ...(!params.tail_lines ? { head_lines: BATCH_SIZE } : {}),
      ...params,
    });
  }, [params, setQuery]);

  const [{ data, error }, startPolling, stopPolling] = usePollingTuringApi(
    apiEndpoint,
    {},
    undefined,
    POLLING_INTERVAL
  );

  useEffect(() => {
    if (isActive) {
      startPolling({ query });
      return stopPolling;
    }
  }, [isActive, query, startPolling, stopPolling]);

  const dispatchData = useCallback(
    (data) => {
      let didCancel = false;
      Promise.resolve(data)
        .then((data) => {
          const { chunk, timestamp } = processLogs(data);

          if (!didCancel && chunk) emitter.emit("data", chunk);

          return timestamp;
        })
        .then((lastTimestamp) => {
          if (!didCancel && lastTimestamp) {
            setQuery((q) => ({
              ...q,
              since_time: lastTimestamp,
              head_lines: BATCH_SIZE,
            }));
          }
        });

      return {
        cancel: () => {
          didCancel = true;
        },
      };
    },
    [emitter, processLogs]
  );

  useEffect(() => {
    if (data && !error) {
      const promise = dispatchData(data);
      return promise.cancel;
    }
  }, [dispatchData, data, error]);

  return { emitter };
};

export default useEventEmitterLogsApi;
