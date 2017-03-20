const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextWebpackPlugin = require('extract-text-webpack-plugin');

module.exports = {
  entry: {
    hypcast: './client/hypcast.js',
  },

  output: {
    filename: '[name].dist.js',
    path: path.resolve('dist/client'),
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
        use: ExtractTextWebpackPlugin.extract({
          fallback: 'style-loader',
          use: [
            { loader: 'css-loader' },
            { loader: 'less-loader' },
          ],
        }),
      },

      {
        test: /\.css$/,
        use: [
          { loader: 'style-loader' },
          {
            loader: 'css-loader',
            options: {
              modules: true,
            },
          },
        ],
      },

      {
        test: /(\.woff2?|\.eot|\.ttf|\.svg)$/,
        use: [
          { loader: 'file-loader' },
        ],
      },

      {
        test: /\.html/,
        use: [
          {
            loader: 'html-loader',
            options: {
              attrs: ['img:src', 'link:href'],
            },
          },
        ],
      },

      {
        test: /\/favicon\.ico$/,
        use: [
          {
            loader: 'file-loader',
            options: {
              name: 'favicon.ico',
            },
          },
        ],
      },
    ],
  },

  plugins: [
    new HtmlWebpackPlugin({
      template: './client/index.html',
    }),
    new ExtractTextWebpackPlugin('hypcast.dist.css'),
  ],
};
