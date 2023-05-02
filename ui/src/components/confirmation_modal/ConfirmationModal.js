import React, { Fragment } from "react";
import { useToggle } from "@caraml-dev/ui-lib";
import { EuiOverlayMask, EuiConfirmModal, EuiProgress } from "@elastic/eui";
import './ConfirmationModal.scss';

export const ConfirmationModal = ({
  title,
  content,
  onCancel = () => {}, // default empty function
  onConfirm,
  confirmButtonText,
  confirmButtonColor,
  isLoading,
  disabled = false,
  ...props
}) => {
  const [isModalVisible, toggleModalVisible] = useToggle();

  return (
    <Fragment>
      {props.children(toggleModalVisible)}

      {isModalVisible && (
        <EuiOverlayMask>
          <EuiConfirmModal
            title={title}
            onCancel={() => {
              onCancel();
              toggleModalVisible();
            }}
            onConfirm={onConfirm}
            cancelButtonText="Cancel"
            confirmButtonText={confirmButtonText}
            buttonColor={confirmButtonColor}
            confirmButtonDisabled={disabled}>
            {content}
            {isLoading && <EuiProgress size="xs" color="accent" />}
          </EuiConfirmModal>
        </EuiOverlayMask>
      )}
    </Fragment>
  );
};
