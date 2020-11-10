const proxy = require("http-proxy-middleware");

module.exports = function(app) {
  app.use(
    "/api/mlp",
    proxy({
      target: process.env.REACT_APP_MLP_API,
      pathRewrite: { "^/api/mlp": "" },
      changeOrigin: true
    })
  );
  app.use(
    "/api/merlin",
    proxy({
      target: process.env.REACT_APP_MERLIN_API,
      pathRewrite: { "^/api/merlin": "" },
      changeOrigin: true
    })
  );
  app.use(
    "/api/turing",
    proxy({
      target: process.env.REACT_APP_TURING_API,
      pathRewrite: { "^/api/turing": "" },
      changeOrigin: true
    })
  );
};
