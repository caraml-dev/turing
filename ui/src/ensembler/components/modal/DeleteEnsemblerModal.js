import React, { useEffect, useRef } from "react";
import { addToast } from "@caraml-dev/ui-lib";
import { useEnsemblerModal } from "./useEnsemblerModal";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../utils/object";

export const DeleteEnsemblerModal = ({
  onSuccess,
  deleteEnsemblerRef,
}) => {
  const closeModalRef = useRef();

  const [ensembler = {}, openModal, closeModal] = useEnsemblerModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] =  useTuringApi(
    `/projects/${ensembler.project_id}/ensemblers/${ensembler.id}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (!isEmpty(ensembler)  && isLoaded && !error) {
      addToast({
        id: `submit-success-delete-${ensembler.id}`,
        title: `Ensembler ${ensembler.name} has been deleted!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, ensembler, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Delete Ensembler"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to delete Ensembler <b>{ensembler.name}</b>
        </p>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteEnsemblerRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
