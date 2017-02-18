const path = require('path');

module.exports = {
  entry: {
    hypcast: './client/hypcast.js',
  },

  output: {
    filename: '[name].dist.js',
    path: path.resolve('client'),
  },

  module: {
    rules: [
      {
        test: /\.js$/,
        use: [
          'babel-loader',
        ],
      },
    ],
  },
};
