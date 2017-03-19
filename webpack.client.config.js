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
        exclude: /node_modules/,
        use: [
          {
            loader: 'babel-loader',
            options: {
              presets: [['es2015', { modules: false }]],
              plugins: ['transform-react-jsx']
            }
          },
        ],
      },
    ],
  },
};
