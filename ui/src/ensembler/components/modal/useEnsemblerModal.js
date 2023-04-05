import { useCallback, useState } from "react";

export const useEnsemblerModal = (closeModalRef) => {
  const [ensembler, setEnsembler] = useState();

  const openModal = useCallback((onSubmit) => {
    return (ensembler) => {
      setEnsembler(ensembler);
      onSubmit();
    };
  }, []);

  const closeModal = useCallback(() => {
    setEnsembler(undefined);
    closeModalRef.current();
  }, [closeModalRef]);

  return [ensembler, openModal, closeModal];
};
