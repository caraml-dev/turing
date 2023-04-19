import React, { useEffect, useRef } from "react";
import { addToast } from "@caraml-dev/mlp-ui";
import { useJobModal } from "./useJobModal";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../utils/object";

export const DeleteJobModal = ({
  onSuccess,
  deleteJobRef,
}) => {
  const closeModalRef = useRef();

  const [job = {}, openModal, closeModal] = useJobModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] =  useTuringApi(
    `/projects/${job.project_id}/jobs/${job.id}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (!isEmpty(job)  && isLoaded && !error) {
      addToast({
        id: `submit-success-terminate-${job.id}`,
        title: `Ensembling Jobs ${job.name} for ensembler ${job.ensembler_id} has been terminated!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, job, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Terminate Ensembling Jobs"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to terminate Ensembling Jobs <b>{job.name}</b>
        </p>
      }
      confirmButtonText="Terminate"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteJobRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
