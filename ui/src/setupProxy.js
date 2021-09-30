const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
  // Proxy settings required for remote components' API calls.
  const remoteProxyPaths = process.env.REMOTE_COMPONENTS_PROXY_PATHS
    ? JSON.parse(process.env.REMOTE_COMPONENTS_PROXY_PATHS)
    : {};
  Object.keys(remoteProxyPaths).forEach((key) => {
    app.use(
      key,
      createProxyMiddleware({
        target: remoteProxyPaths[key],
        pathRewrite: { ["^" + key]: "" },
        changeOrigin: true,
      })
    );
  });

  app.use(
    "/api/mlp",
    createProxyMiddleware({
      target: process.env.REACT_APP_MLP_API,
      pathRewrite: { "^/api/mlp": "" },
      changeOrigin: true,
    })
  );
  app.use(
    "/api/merlin",
    createProxyMiddleware({
      target: process.env.REACT_APP_MERLIN_API,
      pathRewrite: { "^/api/merlin": "" },
      changeOrigin: true,
    })
  );
  app.use(
    "/api/turing",
    createProxyMiddleware({
      target: process.env.REACT_APP_TURING_API,
      pathRewrite: { "^/api/turing": "" },
      changeOrigin: true,
    })
  );
};
