import { useEffect, useState } from "react";

export const useInitiallyLoaded = (isLoaded) => {
  const [hasInitiallyLoaded, setHasInitiallyLoaded] = useState(isLoaded);

  useEffect(() => {
    if (!hasInitiallyLoaded) {
      setHasInitiallyLoaded(isLoaded);
    }
  }, [isLoaded, hasInitiallyLoaded, setHasInitiallyLoaded]);

  return hasInitiallyLoaded;
};
