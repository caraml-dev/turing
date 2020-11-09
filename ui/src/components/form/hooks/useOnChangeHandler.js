import { useCallback, useEffect, useRef } from "react";

export const useOnChangeHandler = onChangeHandler => {
  const nested = useRef({});

  // reset nested handlers when the parent handler is changed
  useEffect(() => {
    nested.current = {};
  }, [onChangeHandler]);

  const onChange = useCallback(
    arg => {
      if (!nested.current[arg]) {
        nested.current[arg] = onChangeHandler.withArg(arg);
      }
      return nested.current[arg];
    },
    [onChangeHandler]
  );

  return { onChangeHandler, onChange };
};
