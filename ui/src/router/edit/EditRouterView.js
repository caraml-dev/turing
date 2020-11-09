import React, { useEffect } from "react";
import { TuringRouter } from "../../services/router/TuringRouter";
import { FormContextProvider } from "../../components/form/context";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { UpdateRouterForm } from "../components/form/UpdateRouterForm";
import { ExperimentEngineContextProvider } from "../../providers/experiments/ExperimentEngineContextProvider";

export const EditRouterView = ({ projectId, router, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: `../`
      },
      {
        text: router.name,
        href: `./`
      },
      {
        text: "Update"
      }
    ]);
  }, [router]);

  return (
    <FormContextProvider data={TuringRouter.fromJson(router)}>
      <ExperimentEngineContextProvider>
        <UpdateRouterForm
          projectId={projectId}
          onCancel={() => props.navigate("../")}
          onSuccess={() => props.navigate("../", { state: { refresh: true } })}
          {...props}
        />
      </ExperimentEngineContextProvider>
    </FormContextProvider>
  );
};
