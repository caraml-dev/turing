/* global __webpack_init_sharing__ */
/* global __webpack_share_scopes__ */
import React, { useMemo } from "react";

// Ref:
// https://github.com/module-federation/module-federation-examples/blob/master/dynamic-system-host
const useRemoteComponent = (scope, module) => {
  const component = useMemo(
    () =>
      React.lazy(async () => {
        // Initializes the shared scope. This fills it with known provided modules from this build and all remotes
        await __webpack_init_sharing__("default");
        const container = window[scope]; // or get the container somewhere else
        // Initialize the container, it may provide shared modules
        await container.init(__webpack_share_scopes__.default);
        const factory = await window[scope].get(module);
        const Module = factory();
        return Module;
      }),
    [scope, module]
  );

  return component;
};

export default useRemoteComponent;
