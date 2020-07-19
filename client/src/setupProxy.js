const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
  app.use(
    createProxyMiddleware("/control-socket", {
      target: "http://localhost:9200",
      ws: true,
      logLevel: "silent",
    }),
  );
};
