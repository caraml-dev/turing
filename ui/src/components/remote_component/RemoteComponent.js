import React from "react";
import useRemoteComponent from "../../hooks/useRemoteComponent";

export const RemoteComponent = ({ scope, name, fallback, ...props }) => {
  const Component = useRemoteComponent(scope, name);
  return (
    <React.Suspense fallback={fallback}>
      <Component {...props} />
    </React.Suspense>
  );
};
