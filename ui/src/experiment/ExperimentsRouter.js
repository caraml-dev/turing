import React from "react";

export const ExperimentsRouter = ({ projectId }) => {
  const View = React.lazy(() => import("expEngine/ExperimentsLandingPage"));

  return (
    <React.Suspense fallback="Loading Experiments">
      <View projectId={projectId} />
    </React.Suspense>
  );
};
