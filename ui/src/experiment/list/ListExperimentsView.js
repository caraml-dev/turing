import React from "react";

export const ListExperimentsView = ({ projectId, ...props }) => {
  const View = React.lazy(() => import("expEngine/ListExperimentsView"));

  return (
    <React.Suspense fallback="Loading Button">
      <View projectId={projectId} props={props} />
    </React.Suspense>
  );
};
