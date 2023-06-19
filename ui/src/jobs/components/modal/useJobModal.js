import { useCallback, useState } from "react";

export const useJobModal = (closeModalRef) => {
  const [job, setJob] = useState();

  const openModal = useCallback((onSubmit) => {
    return (job) => {
      setJob(job);
      onSubmit();
    };
  }, []);

  const closeModal = useCallback(() => {
    setJob(undefined);
    closeModalRef.current();
  }, [closeModalRef]);

  return [job, openModal, closeModal];
};
