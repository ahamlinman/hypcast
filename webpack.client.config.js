const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

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

      {
        test: /\.less$/,
        use: [
          { loader: 'style-loader' },
          { loader: 'css-loader' },
          { loader: 'less-loader' },
        ],
      },

      {
        test: /(\.woff2?|\.eot|\.ttf|\.svg)$/,
        use: [
          { loader: 'file-loader' },
        ],
      },
    ],
  },

  plugins: [
    new HtmlWebpackPlugin({
      template: './client/index.ejs',
    }),
  ],
};
