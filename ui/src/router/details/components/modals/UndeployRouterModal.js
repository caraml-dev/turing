import React, { useEffect, useRef } from "react";
import { addToast } from "@gojek/mlp-ui";
import { useTuringApi } from "../../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../../utils/object";
import { useRouterModal } from "./useRouterModal";

export const UndeployRouterModal = ({ onSuccess, undeployRouterRef }) => {
  const closeModalRef = useRef();
  const [router = {}, openModal, closeModal] = useRouterModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useTuringApi(
    `/projects/${router.project_id}/routers/${router.id}/undeploy`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (!isEmpty(router) && isLoaded && !error) {
      addToast({
        id: `submit-success-undeploy-${router.name}`,
        title: `Router ${router.name} has been undeployed!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, router, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Undeploy Turing Router"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to undeploy Router <b>{router.name}</b> from{" "}
          <b>{router.environment_name}</b> environment.
        </p>
      }
      confirmButtonText="Undeploy"
      confirmButtonColor="danger"
    >
      {(onSubmit) =>
        (undeployRouterRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
