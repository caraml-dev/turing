import React, { useEffect, useRef, useState } from "react";
import { addToast } from "@caraml-dev/ui-lib";
import { useJobModal } from "./useJobModal";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../utils/object";
import { EuiFieldText } from "@elastic/eui";
import { isActiveJobStatus } from "../../../services/job/JobStatus";

export const DeleteJobModal = ({
  onSuccess,
  deleteJobRef,
}) => {
  const closeModalRef = useRef();
  const [deleteConfirmation, setDeleteConfirmation] = useState('')

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
        title: `Ensembling Job ${job.name} for ensembler ${job.ensembler_id} has been terminated!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, job, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title={isActiveJobStatus(job.status) ? "Terminate Ensembling Jobs" : "Delete Ensembling Jobs" }
      onCancel={() => setDeleteConfirmation("")}
      onConfirm={(arg) => {submitForm(arg); setDeleteConfirmation("")}}
      isLoading={isLoading}
      disabled={deleteConfirmation !== job.name}
      content={
        <div>
          <p>
            You are about to terminate Ensembling Jobs <b>{job.name}</b>
          </p>
          To confirm, please type "<b>{job.name}</b>" in the box below
          <EuiFieldText     
            fullWidth            
            placeholder={job.name}
            value={deleteConfirmation}
            onChange={(e) => setDeleteConfirmation(e.target.value)}
            isInvalid={deleteConfirmation !== job.name} />   
        </div>
      }
      confirmButtonText={isActiveJobStatus(job.status) ? "Terminate" : "Delete"}
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteJobRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
