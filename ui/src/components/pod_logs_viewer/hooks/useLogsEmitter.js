import { useCallback, useEffect, useState } from "react";
import { useTuringPollingApiEmitter } from "../../../hooks/useTuringPollingApiEmitter";
import { useLogsApiEmitter } from "./useLogsApiEmitter";

export const useLogsEmitter = (
  apiEndpoint,
  query,
  extractTimestamp,
  processLogs,
  configOptions
) => {
  const [apiOptions, setApiOptions] = useState({ query });

  const setQuery = useCallback(
    (setQuery) => {
      setApiOptions((options) => ({
        ...options,
        query: setQuery(options.query),
      }));
    },
    [setApiOptions]
  );

  useEffect(() => {
    setQuery(() => query);
  }, [query, setQuery]);

  const { emitter: apiEmitter } = useTuringPollingApiEmitter(
    apiEndpoint,
    apiOptions,
    configOptions.pollInterval
  );

  useEffect(() => {
    apiEmitter.on("data", (entries) => {
      const lastTimestamp = extractTimestamp(entries);

      if (!!lastTimestamp) {
        setQuery((q) => ({
          ...q,
          since_time: lastTimestamp,
          head_lines: configOptions.batchSize,
        }));
      }
    });

    apiEmitter.emit("start");

    return () => {
      apiEmitter.emit("abort");
    };
  }, [apiEmitter, setQuery, extractTimestamp, configOptions.batchSize]);

  const { emitter } = useLogsApiEmitter(apiEmitter, processLogs);

  return { emitter };
};
