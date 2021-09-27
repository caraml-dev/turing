import React, { useCallback, useState, useMemo } from "react";
import { set } from "./utils";
import { StackableFunction } from "./functions/stackable_function";
import { useOnChangeHandler } from "./hooks/useOnChangeHandler";

export const FormContext = React.createContext({});

export const FormContextProvider = ({ data: initData, ...props }) => {
  const [data, setData] = useState(initData);

  const handleChanges = useCallback(
    (paths, value) => {
      const path = paths.filter((part) => !!part).join(".");
      setData((data) => {
        set(data, path, value);
        return Object.assign(Object.create(data), data);
      });
    },
    [setData]
  );

  const rootHandler = useMemo(
    () => new StackableFunction([], handleChanges),
    [handleChanges]
  );
  const { onChangeHandler, onChange } = useOnChangeHandler(rootHandler);

  return (
    <FormContext.Provider value={{ data, onChange, onChangeHandler }}>
      {props.children}
    </FormContext.Provider>
  );
};
