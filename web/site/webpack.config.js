module.exports = {
    mode: 'production',

    performance: {
        maxAssetSize: 1048576,
        maxEntrypointSize: 1048576,
    },

    resolve: {
        extensions: ['*', '.mjs', '.js', '.json']
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