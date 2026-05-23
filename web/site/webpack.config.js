const path = require('path');
const TerserPlugin = require('terser-webpack-plugin');

module.exports = {
    mode: 'production',

    entry: {
        visualize: './js/visualize.js',
        energyViewer: './js/energyViewer.js',
        statsViewer: './js/statsViewer.js',
    },

    output: {
        path: path.resolve(__dirname, 'static', 'js'),
        filename: '[name].js',
    },

    optimization: {
        // Keep license comments inside the bundle instead of emitting separate
        // *.LICENSE.txt files, which go-bindata would otherwise embed and serve.
        minimizer: [new TerserPlugin({extractComments: false})],
    },

    performance: {
        maxAssetSize: 1048576,
        maxEntrypointSize: 1048576,
    },

    resolve: {
        extensions: ['.mjs', '.js', '.json'],
    },

    module: {
        rules: [
            {
                test: /\.mjs$/,
                include: /node_modules/,
                type: 'javascript/auto'
            }
        ]
    }
};
