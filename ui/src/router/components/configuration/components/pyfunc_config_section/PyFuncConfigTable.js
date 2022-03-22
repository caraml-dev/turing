import React, { useContext } from "react";
import { EuiDescriptionList } from "@elastic/eui";
import EnsemblersContext from "../../../../../providers/ensemblers/context";

export const PyFuncConfigTable = ({ config: { ensembler_id } }) => {
  const { ensemblers } = useContext(EnsemblersContext);

  const ensembler_name = Object.values(ensemblers)
    .filter((value) => value.id === ensembler_id)
    .map((ensembler) => ensembler.name);

  const items = [
    {
      title: "Ensembler ID",
      description: ensembler_id,
    },
    {
      title: "Ensembler Name",
      description: ensembler_name,
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  );
};
