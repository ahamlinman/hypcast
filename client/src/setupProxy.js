const { createProxyMiddleware } = require("http-proxy-middleware");

const defaultOptions = { target: "http://localhost:9200", logLevel: "silent" };

module.exports = function (app) {
  app.use(createProxyMiddleware("/config", { ...defaultOptions }));
  app.use(
    createProxyMiddleware("/control-socket", {
      ...defaultOptions,
      ws: true,
    }),
  );
};
