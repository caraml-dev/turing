import { useCallback, useMemo } from "react";
import mitt from "mitt";

export const useLogsApiEmitter = (dataEmitter, processLogs) => {
  const dispatchData = useCallback(
    (data, emitter) => {
      let didCancel = false;
      Promise.resolve(data)
        .then(processLogs)
        .then((entries) => {
          const chunk = entries.join("\n");

          if (!didCancel && !!chunk) {
            emitter.emit("data", chunk);
            if (!chunk.endsWith("\n")) emitter.emit("data", "\n");
          }
        });

      return {
        cancel: () => {
          didCancel = true;
        },
      };
    },
    [processLogs]
  );

  const emitter = useMemo(() => {
    const emitter = new mitt();

    dataEmitter.on("data", (data) => dispatchData(data, emitter));

    emitter.on("start", () => {
      dataEmitter.emit("start");
      emitter.emit("data", "Fetching logs...\n");
    });

    emitter.on("abort", () => {
      dataEmitter.emit("stop");
    });

    return Object.freeze({
      on: (event, fn) => emitter.on(event, fn),
      off: (event, fn) => emitter.off(event, fn),
      emit: (event, payload) => emitter.emit(event, payload),
    });
  }, [dataEmitter, dispatchData]);

  return { emitter };
};
