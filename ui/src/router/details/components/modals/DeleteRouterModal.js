import React, { useEffect, useRef } from "react";
import { addToast } from "@gojek/mlp-ui";
import { useTuringApi } from "../../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../../utils/object";
import { useRouterModal } from "./useRouterModal";

export const DeleteRouterModal = ({ onSuccess, deleteRouterRef }) => {
  const closeModalRef = useRef();
  const [router = {}, openModal, closeModal] = useRouterModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useTuringApi(
    `/projects/${router.project_id}/routers/${router.id}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (!isEmpty(router) && isLoaded && !error) {
      addToast({
        id: `submit-success-delete-${router.name}`,
        title: `Router ${router.name} has been deleted!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, router, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Delete Turing Router"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to delete Router <b>{router.name}</b> and all its
          versions.
        </p>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteRouterRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
