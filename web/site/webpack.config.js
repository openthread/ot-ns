module.exports = {
    mode: 'production',

    performance: {
        maxAssetSize: 1048576,
        maxEntrypointSize: 1048576,
    },

    optimization: {
        moduleIds: 'hashed',
        chunkIds: 'named'
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