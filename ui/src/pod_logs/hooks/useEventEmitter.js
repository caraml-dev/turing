import { useMemo, useState } from "react";
import mitt from "mitt";

const useEventEmitter = () => {
  const [doPoll, setDoPoll] = useState(false);

  const emitter = useMemo(() => {
    const emitter = new mitt();

    emitter.on("start", () => {
      emitter.emit("data", "Fetching logs...\n");
      setDoPoll(true);
    });

    emitter.on("abort", () => {
      setDoPoll(false);
    });

    return Object.freeze({
      on: (event, fn) => emitter.on(event, fn),
      off: (event, fn) => emitter.off(event, fn),
      emit: (event, payload) => emitter.emit(event, payload),
    });
  }, [setDoPoll]);

  return { emitter, isActive: doPoll };
};

export default useEventEmitter;
