import React, { Fragment, useCallback, useRef } from "react";
import { UndeployRouterModal } from "./modals/UndeployRouterModal";
import { RedeployRouterModal } from "./modals/RedeployRouterModal";
import { DeleteRouterModal } from "./modals/DeleteRouterModal";
import { Status } from "../../../services/status/Status";

export const RouterActions = ({
  onEditRouter,
  onDeploySuccess,
  onUndeploySuccess,
  onDeleteSuccess,
  children
}) => {
  const undeployRouterRef = useRef();
  const redeployRouterRef = useRef();
  const deleteRouterRef = useRef();

  const actions = useCallback(
    router => {
      const status = Status.fromValue(router.status);
      return [
        {
          name: "Edit Router",
          icon: "documentEdit",
          disabled: status === Status.PENDING,
          onClick: onEditRouter
        },
        {
          name: "Undeploy Router",
          icon: "exportAction",
          disabled: status === Status.PENDING,
          hidden: status === Status.UNDEPLOYED,
          onClick: () => undeployRouterRef.current(router)
        },
        {
          name: "Redeploy Router",
          icon: "importAction",
          hidden: [Status.DEPLOYED, Status.PENDING].includes(status),
          onClick: () => redeployRouterRef.current(router)
        },
        {
          name: "Delete Router",
          icon: "trash",
          color: "danger",
          disabled: [Status.DEPLOYED, Status.PENDING].includes(status),
          onClick: () => deleteRouterRef.current(router)
        }
      ];
    },
    [onEditRouter]
  );

  return (
    <Fragment>
      <RedeployRouterModal
        onSuccess={onDeploySuccess}
        redeployRouterRef={redeployRouterRef}
      />
      <UndeployRouterModal
        onSuccess={onUndeploySuccess}
        undeployRouterRef={undeployRouterRef}
      />
      <DeleteRouterModal
        onSuccess={onDeleteSuccess}
        deleteRouterRef={deleteRouterRef}
      />
      {children(actions)}
    </Fragment>
  );
};
