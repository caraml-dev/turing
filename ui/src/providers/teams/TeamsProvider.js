import React from "react";
import TeamsContext from "./context";
import { useMerlinApi } from "../../hooks/useMerlinApi";

export const TeamsProvider = ({ ...props }) => {
  const [{ data: teams }] = useMerlinApi(
    `/alerts/teams`,
    { muteError: true },
    []
  );

  return (
    <TeamsContext.Provider value={teams}>
      {props.children}
    </TeamsContext.Provider>
  );
};
