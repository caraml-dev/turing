import { useContext, useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import EnsemblersContext from "../../../providers/ensemblers/context";

export const EnsemblingJobConfigView = ({ job }) => {
  const ensemblers = useContext(EnsemblersContext);

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Jobs",
        href: "../",
      },
      {
        text: `Job ${job.id}`,
        href: "./",
      },
      {
        text: "Details",
      },
    ]);
  }, [job.id]);

  return null;
};
