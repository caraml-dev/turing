import React, { useEffect, useRef } from "react";
import { addToast } from "@caraml-dev/ui-lib";
import { useTuringApi } from "../../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../../components/confirmation_modal/ConfirmationModal";
import { useVersionModal } from "./useVersionModal";

export const DeleteVersionModal = ({
  projectId,
  routerId,
  routerName,
  onSuccess,
  deleteVersionRef,
}) => {
  const closeModalRef = useRef();

  const [version, openModal, closeModal] = useVersionModal(closeModalRef);
  const [{ isLoading, isLoaded, error }, submitForm] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/versions/${version}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (version && isLoaded && !error) {
      addToast({
        id: `submit-success-delete-${version}`,
        title: `Router version ${version} has been deleted!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, version, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Delete Turing Router Version"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to delete Router <b>{routerName}</b> Version{" "}
          <b>{version}</b>.
        </p>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteVersionRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
