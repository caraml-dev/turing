import React, { useEffect, useRef } from "react";
import { addToast } from "@gojek/mlp-ui";
import { useTuringApi } from "../../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../../utils/object";
import { useRouterModal } from "./useRouterModal";

export const RedeployRouterModal = ({ onSuccess, redeployRouterRef }) => {
  const closeModalRef = useRef();
  const [router = {}, openModal, closeModal] = useRouterModal(closeModalRef);

  const [{ data, isLoading, isLoaded, error }, submitForm] = useTuringApi(
    `/projects/${router.project_id}/routers/${router.id}/deploy`,
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
        id: `submit-success-redeploy-${router.name}`,
        title: `Router ${router.name} version ${data.version} has been redeployed!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [data, isLoaded, error, router, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Redeploy Turing Router"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to redeploy Router <b>{router.name}</b>. The most
          recently deployed version will be redeployed.
        </p>
      }
      confirmButtonText="Redeploy"
      confirmButtonColor="primary">
      {(onSubmit) =>
        (redeployRouterRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
