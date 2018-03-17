const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

function getStyleLoaders(mode, extOptions = {}) {
  const options = Object.assign(
    { importLoaders: 1 },
    (mode === 'production') ? { minimize: true } : {},
    extOptions,
  );

  return [
    { loader: MiniCssExtractPlugin.loader },
    { loader: 'css-loader', options },
    { loader: 'less-loader' },
  ];
}

module.exports = function(_, argv) {
  const mode = (argv.mode || 'development');

  return {
    entry: { hypcast: './client/hypcast.jsx' },

    output: {
      filename: '[name].dist.js',
      path: path.resolve(__dirname, 'dist', 'client'),
    },

    module: {
      rules: [
        {
          test: /\.jsx?$/,
          exclude: /node_modules/,
          use: [
            {
              loader: 'babel-loader',
              options: {
                plugins: ['transform-react-jsx'],
              },
            },
          ],
        },

        {
          test: /\.less$/,
          oneOf: [
            {
              include: path.resolve(__dirname, 'client', 'ui'),
              use: getStyleLoaders(mode, { modules: true }),
            },
            { use: getStyleLoaders(mode) },
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
              options: { attrs: ['img:src', 'link:href'] },
            },
          ],
        },

        {
          test: /\/favicon\.ico$/,
          use: [
            {
              loader: 'file-loader',
              options: { name: 'favicon.ico' },
            },
          ],
        },
      ],
    },

    resolve: { extensions: ['.js', '.json', '.jsx'] },

    plugins: [
      new HtmlWebpackPlugin({ template: './client/index.html' }),
      new MiniCssExtractPlugin({ filename: 'hypcast.dist.css' }),
    ],
  };
}
