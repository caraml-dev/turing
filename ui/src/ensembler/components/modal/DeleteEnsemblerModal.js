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

  const [canDeleteEnsembler, setCanDeleteEnsembler] = useState(true)

  const [ensemblerUsedByActiveRouterVersion, setEnsemblerUsedByActiveRouterVersion] = useState(false)
  const [ensemblerUsedByActiveEnsemblingJob, setEnsemblerUsedByActiveEnsemblingJob] = useState(false)
  const [ensemblerUsedByCurrentRouterVersion, setEnsemblerUsedByCurrentRouterVersion] = useState(false)

  const [deleteConfirmation, setDeleteConfirmation] = useState('')
  const [ensembler = {}, openModal, closeModal] = useEnsemblerModal(closeModalRef);

  useEffect(() => {
    // if ensembler is used by one of the component, immediately set can delete ensembler to false
    setCanDeleteEnsembler(!(ensemblerUsedByActiveEnsemblingJob || ensemblerUsedByActiveRouterVersion || ensemblerUsedByCurrentRouterVersion))
  }, [ensemblerUsedByActiveEnsemblingJob, ensemblerUsedByActiveRouterVersion, ensemblerUsedByCurrentRouterVersion])

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

  const modalClosed = () => {
    setCanDeleteEnsembler(true)
    setEnsemblerUsedByCurrentRouterVersion(false)
    setDeleteConfirmation("")
  }

  return (
    <ConfirmationModal
      title="Delete Ensembler"
      onCancel={() => modalClosed()}
      onConfirm={(arg) => {submitForm(arg); modalClosed()}}
      isLoading={isLoading}
      content={
        <div>
          {canDeleteEnsembler ? (
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
          ) : ensemblerUsedByCurrentRouterVersion ? (
            <div>
              You cannot delete this ensembler because it is associated with a router version that is currently being used by a router
              <br/> <br/> If you still wish to delete this ensembler, please <b>Deploy</b> another version on this router.
            </div>
          ) : (
            <div>
              You cannot delete this ensembler because there are <b>Active Router Versions</b> or <b>Ensembling Jobs</b> that use this ensembler. 
              <br/> <br/> If you still wish to delete this ensembler, please <b>Undeploy</b> router versions and <b>Terminate</b> ensembling jobs that use this ensembler.
            </div> 
          )}
          {/* Only show The Ensembling Table if ensembler is not used by current router version */}
          {!ensemblerUsedByCurrentRouterVersion && (
            <ListEnsemblingJobsForEnsemblerTable 
              projectID={ensembler.project_id}
              ensemblerID={ensembler.id}
              canDeleteEnsembler={canDeleteEnsembler}
              setEnsemblerUsedByActiveEnsemblingJob={setEnsemblerUsedByActiveEnsemblingJob}
            />   
          )}
          <ListRouterVersionsForEnsemblerTable 
            projectID={ensembler.project_id}
            ensemblerID={ensembler.id}
            canDeleteEnsembler={canDeleteEnsembler}
            setEnsemblerUsedByActiveRouterVersion={setEnsemblerUsedByActiveRouterVersion}
            setEnsemblerUsedByCurrentRouterVersion={setEnsemblerUsedByCurrentRouterVersion}
            ensemblerUsedByCurrentRouterVersion={ensemblerUsedByCurrentRouterVersion}
          />
        </div>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger"
      disabled={!canDeleteEnsembler || deleteConfirmation !== ensembler.name}>
      {(onSubmit) =>
        (deleteEnsemblerRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
