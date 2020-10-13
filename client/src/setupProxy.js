const { createProxyMiddleware } = require("http-proxy-middleware");

const proxyOptions = {
  target: "http://localhost:9200",
  logLevel: "silent",
  ws: true,
};

module.exports = function (app) {
  app.use(createProxyMiddleware("/api", proxyOptions));

  // TODO: Legacy routes, remove after /api routes are finished.
  app.use(createProxyMiddleware("/config", proxyOptions));
  app.use(createProxyMiddleware("/old-control-socket", proxyOptions));
};
