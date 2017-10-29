const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextWebpackPlugin = require('extract-text-webpack-plugin');

function extractStylesPlugin(options = {}) {
  return ExtractTextWebpackPlugin.extract({
    fallback: 'style-loader',
    use: [
      {
        loader: 'css-loader',
        options: Object.assign({ importLoaders: 1 }, options),
      },
      { loader: 'less-loader' },
    ],
  });
}

module.exports = {
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
              presets: [['es2015', { modules: false }]],
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
            use: extractStylesPlugin({ modules: true }),
          },
          { use: extractStylesPlugin() },
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
    new ExtractTextWebpackPlugin('hypcast.dist.css'),
  ],
};
