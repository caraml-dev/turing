import { useCallback, useEffect, useState } from "react";
import { usePollingTuringApi } from "../../../hooks/usePollingTuringApi";
import useEventEmitter from "./useEventEmitter";

const POLLING_INTERVAL = 5000;
const BATCH_SIZE = 500;

const useEventEmitterLogsApi = (projectId, routerId, params, formatMessage) => {
  const { emitter, isActive } = useEventEmitter();

  const [query, setQuery] = useState();

  useEffect(() => {
    setQuery({
      head_lines: !params.tail_lines ? BATCH_SIZE : "",
      ...params
    });
  }, [params, setQuery]);

  const [{ data, error }, startPolling, stopPolling] = usePollingTuringApi(
    `/projects/${projectId}/routers/${routerId}/logs`,
    {},
    [],
    POLLING_INTERVAL
  );

  useEffect(() => {
    if (isActive) {
      startPolling({ query });
      return stopPolling;
    }
  }, [isActive, query, startPolling, stopPolling]);

  const dispatchData = useCallback(
    data => {
      let didCancel = false;
      Promise.resolve(data)
        .then(entries => {
          let logChunk = entries.map(formatMessage).join("\n");
          logChunk = logChunk.endsWith("\n") ? logChunk : `${logChunk}\n`;

          if (!didCancel) emitter.emit("data", logChunk);

          return entries[entries.length - 1].timestamp;
        })
        .then(lastTimestamp => {
          if (!didCancel) {
            setQuery(q => ({
              ...q,
              since_time: lastTimestamp,
              head_lines: BATCH_SIZE
            }));
          }
        });

      return {
        cancel: () => {
          didCancel = true;
        }
      };
    },
    [emitter, formatMessage]
  );

  useEffect(() => {
    if (data.length && !error) {
      const promise = dispatchData(data);
      return promise.cancel;
    }
  }, [dispatchData, data, error]);

  return { emitter };
};

export default useEventEmitterLogsApi;