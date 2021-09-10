import { useCallback, useState } from "react";

export const useVersionModal = (closeModalRef) => {
  const [version, setVersion] = useState();

  const openModal = useCallback((onSubmit) => {
    return (version) => {
      setVersion(version);
      onSubmit();
    };
  }, []);

  const closeModal = useCallback(() => {
    setVersion(undefined);
    closeModalRef.current();
  }, [closeModalRef]);

  return [version, openModal, closeModal];
};
