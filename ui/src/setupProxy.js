const proxy = require("http-proxy-middleware");

let expEnginePathRewrite = {};
expEnginePathRewrite[
  `^${process.env.REACT_APP_DEFAULT_EXPERIMENT_ENGINE_UNROUTABLE_PATH}`
] = "";

module.exports = function (app) {
  app.use(
    "/api/mlp",
    proxy({
      target: process.env.REACT_APP_MLP_API,
      pathRewrite: { "^/api/mlp": "" },
      changeOrigin: true,
    })
  );
  app.use(
    "/api/merlin",
    proxy({
      target: process.env.REACT_APP_MERLIN_API,
      pathRewrite: { "^/api/merlin": "" },
      changeOrigin: true,
    })
  );
  app.use(
    "/api/turing",
    proxy({
      target: process.env.REACT_APP_TURING_API,
      pathRewrite: { "^/api/turing": "" },
      changeOrigin: true,
    })
  );
  /* The experiment engine is expected to use "/api/exp" as the API path on local env,
     to bypass CORS */
  app.use(
    process.env.REACT_APP_DEFAULT_EXPERIMENT_ENGINE_UNROUTABLE_PATH,
    proxy({
      target: process.env.REACT_APP_DEFAULT_EXPERIMENT_ENGINE_API_HOST,
      pathRewrite: expEnginePathRewrite,
      changeOrigin: true,
    })
  );
};
