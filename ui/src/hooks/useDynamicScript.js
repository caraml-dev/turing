import React, { useEffect } from "react";

// Ref:
// https://github.com/module-federation/module-federation-examples/blob/master/dynamic-system-host
const useDynamicScript = (args) => {
  const [ready, setReady] = React.useState(false);
  const [failed, setFailed] = React.useState(false);

  React.useEffect(() => {
    const element = document.createElement("script");

    // Append query string to prevent serving the cached version of the file
    element.src = args.url + "?" + Math.floor(Date.now() / 1000);
    element.type = "text/javascript";
    element.async = true;

    setReady(false);
    setFailed(false);

    element.onload = () => {
      setReady(true);
    };

    element.onerror = () => {
      console.error(`Dynamic Script Error: ${args.url}`);
      setReady(false);
      setFailed(true);
    };

    document.head.appendChild(element);

    return () => {
      document.head.removeChild(element);
    };
  }, [args.url]);

  return {
    ready,
    failed,
  };
};

// Renderless component wrapper
export const LoadDynamicScript = (props) => {
  const { ready, failed } = useDynamicScript({
    url: props.url,
  });

  useEffect(() => {
    if (props.url) {
      props.setConfigStatusReady(ready);
      props.setConfigStatusFailed(failed);
      props.setConfigStatusLoaded(true);
    }
  }, [props, ready, failed]);

  return null;
};

export default useDynamicScript;
