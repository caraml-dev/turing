import React, { Fragment, useCallback, useMemo, useRef } from "react";
import { VersionDeploymentModal } from "./modals/VersionDeploymentModal";
import { DeleteVersionModal } from "./modals/DeleteVersionModal";
import { Status } from "../../../services/status/Status";

export const RouterVersionActions = ({
  router,
  onDeploySuccess,
  onDeleteSuccess,
  children
}) => {
  const deployVersionRef = useRef();
  const deleteVersionRef = useRef();

  const routerStatus = useMemo(() => Status.fromValue(router.status), [
    router.status
  ]);

  const configStatus = config => Status.fromValue(config.status);

  const isActiveConfig = useCallback(
    config => config.version === (router.config || {}).version,
    [router.config]
  );

  const actions = useMemo(
    () => [
      {
        label: "Deploy",
        name: "Deploy this version",
        icon: "importAction",
        available: config =>
          !(isActiveConfig(config) && routerStatus !== Status.UNDEPLOYED),
        enabled: () => routerStatus !== Status.PENDING,
        onClick: config => deployVersionRef.current(config.version)
      },
      {
        label: "Delete",
        name: "Delete this version",
        icon: "trash",
        color: "danger",
        available: () => true,
        enabled: config =>
          !isActiveConfig(config) && configStatus(config) !== Status.PENDING,
        onClick: config => deleteVersionRef.current(config.version)
      }
    ],
    [routerStatus, isActiveConfig]
  );

  return (
    <Fragment>
      <VersionDeploymentModal
        projectId={router.project_id}
        routerId={router.id}
        onSuccess={onDeploySuccess}
        deployVersionRef={deployVersionRef}
      />
      <DeleteVersionModal
        projectId={router.project_id}
        routerId={router.id}
        routerName={router.name}
        onSuccess={onDeleteSuccess}
        deleteVersionRef={deleteVersionRef}
      />
      {children(actions)}
    </Fragment>
  );
};
