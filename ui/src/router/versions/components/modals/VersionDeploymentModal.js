import React, { useEffect, useRef } from "react";
import { addToast } from "@gojek/mlp-ui";
import { useTuringApi } from "../../../../hooks/useTuringApi";
import { ConfirmationModal } from "../../../../components/confirmation_modal/ConfirmationModal";
import { useVersionModal } from "./useVersionModal";

export const VersionDeploymentModal = ({
  projectId,
  routerId,
  onSuccess,
  deployVersionRef
}) => {
  const closeModalRef = useRef();

  const [version, openModal, closeModal] = useVersionModal(closeModalRef);
  const [{ isLoading, isLoaded, error }, submitForm] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/versions/${version}/deploy`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" }
    },
    {},
    false
  );

  useEffect(() => {
    if (version && isLoaded && !error) {
      addToast({
        id: `submit-success-deploy-${version}`,
        title: `Router version ${version} is deployed!`,
        color: "success",
        iconType: "check"
      });
      onSuccess(version);
      closeModal();
    }
  }, [isLoaded, error, version, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Deploy Turing Router"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={<p>You are about to deploy Router version {version}.</p>}
      confirmButtonText="Deploy"
      confirmButtonColor="primary">
      {onSubmit =>
        (deployVersionRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
