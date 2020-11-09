import { useCallback, useState } from "react";

export const useRouterModal = closeModalRef => {
  const [router, setRouter] = useState();

  const openModal = useCallback(onSubmit => {
    return router => {
      setRouter(router);
      onSubmit();
    };
  }, []);

  const closeModal = useCallback(() => {
    setRouter(undefined);
    closeModalRef.current();
  }, [closeModalRef]);

  return [router, openModal, closeModal];
};
