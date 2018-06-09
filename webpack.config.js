const path = require('path');
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

module.exports = function webpackConfig(_, argv) {
  const mode = (argv.mode || 'development');

  return {
    entry: { hypcast: './client/hypcast.tsx' },

    output: {
      filename: '[name].dist.js',
      path: path.resolve(__dirname, 'dist', 'client'),
    },

    module: {
      rules: [
        {
          test: /\.tsx?$/,
          exclude: /node_modules/,
          use: [
            {
              loader: 'awesome-typescript-loader',
              options: { configFileName: './client/tsconfig.json' },
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
              options: Object.assign(
                { attrs: ['img:src', 'link:href'] },
                (mode === 'production') ? { minimize: true } : {},
              ),
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

    resolve: { extensions: ['.ts', '.tsx', '.js', '.jsx', '.json'] },

    plugins: [
      new HtmlWebpackPlugin({ template: './client/index.html' }),
      new MiniCssExtractPlugin({ filename: '[name].dist.css' }),
    ],
  };
};
