var path = require('path');
var webpack = require('webpack');

module.exports = {
  entry: './js/app.js',
  output: { path: __dirname, filename: 'assets/bundle.js' },
  module: {
    loaders: [
      {
        test: /.jsx?$/,
        loader: 'babel-loader',
        exclude: /node_modules/,
        query: {
          presets: ['es2015', 'react']
        }
      },
        {
            test: /\.scss$/,
            loaders: ['style', 'css', 'sass']
        }
    ]
  },
  node: {
    fs: 'empty',
    net: 'empty',
    tls: 'empty'
  },
  resolve: {
    extensions: ['', '.js', '.jsx']
  }
};
