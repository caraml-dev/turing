import React, { useEffect, useRef, useState } from "react";
import { addToast } from "@caraml-dev/ui-lib";
import { useEnsemblerModal } from "./useEnsemblerModal";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../components/confirmation_modal/ConfirmationModal";
import { isEmpty } from "../../../utils/object";
import { ListEnsemblingJobsForEnsemblerTable } from "../table/ListEnsemblingJobsForEnsemblerTable";
import { ListRouterVersionsForEnsemblerTable } from "../table/ListRouterVersionsForEnsemblerTable";
import { EuiFieldText } from "@elastic/eui";

export const DeleteEnsemblerModal = ({
  onSuccess,
  deleteEnsemblerRef,
}) => {
  const closeModalRef = useRef();

  const [disablePopup, setDisablePopup] = useState(false)
  const [deleteConfirmation, setDeleteConfirmation] = useState('')
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

  const updateStatus = (newStatus) => {
    if ((disablePopup && newStatus) || (!disablePopup && !newStatus)) {
      // If the current status and the new status are the same, do nothing.
      return;
    } else {
      setDisablePopup(true);
    }
  };

  const modalClosed = () => {
    setDisablePopup(false)
    setDeleteConfirmation("")
  }

  return (
    <ConfirmationModal
      title="Delete Ensembler"
      onCancel={() => modalClosed()}
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <div>
          {disablePopup ? (
            <div>
              <p>
              You cannot delete this ensembler because there are <b>Active Router Versions</b> or <b>Ensembling Jobs</b> that use this ensembler. 
              If you still wish to delete this ensembler, please undeploy any router versions that use it or terminate any ensembling jobs that use it.
              </p>
            </div> 
          ) : (
            <div>
              <p>
              You are about to delete Ensembler <b>{ensembler.name}</b>. This action cannot be undone. 
              </p>
              To confirm, please type "<b>{ensembler.name}</b>" in the box below
              <EuiFieldText     
                fullWidth            
                placeholder={ensembler.name}
                value={deleteConfirmation}
                onChange={(e) => setDeleteConfirmation(e.target.value)}
                isInvalid={deleteConfirmation !== ensembler.name} />              
            </div>
              
            
          )}
          <ListEnsemblingJobsForEnsemblerTable 
            projectID={ensembler.project_id}
            ensemblerID={ensembler.id}
            setDisablePopup={updateStatus}
            disablePopup={disablePopup}
          />
          <br/>
          <ListRouterVersionsForEnsemblerTable 
            projectID={ensembler.project_id}
            ensemblerID={ensembler.id}
            setDisablePopup={updateStatus}
            disablePopup={disablePopup}
          />
        </div>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger"
      disabled={disablePopup || deleteConfirmation !== ensembler.name}>
      {(onSubmit) =>
        (deleteEnsemblerRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
