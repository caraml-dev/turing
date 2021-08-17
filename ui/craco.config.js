const { ModuleFederationPlugin } = require("webpack").container;
// Remove source code dependency react-lazylog which causes problems with sharing
const { ["react-lazylog"]: undefined, ...sharedDeps } = require("./package.json").dependencies;

module.exports = ({ env }) => ({
    reactScriptsVersion: "webpack-5-react-scripts",
    plugins: [
        {
            plugin: {
                overrideWebpackConfig: ({ webpackConfig, cracoConfig, pluginOptions, context: { env, paths } }) => {
                    return env == "development" ?
                        {
                            ...webpackConfig,
                            // default is optimization.splitChunks.name = true for dev which currently yields an error.
                            // Ref: https://github.com/webpack/webpack-cli/issues/2003
                            optimization: {
                                ...webpackConfig.optimization,
                                splitChunks: {
                                    name: false,
                                },
                            },
                        } : webpackConfig;
                },
            },
        },
    ],
    webpack: {
        plugins: {
            add: [
                new ModuleFederationPlugin({
                    name: "turing",
                    remotes: {
                        expEngine: "xp@http://localhost:3002/xp/remoteEntry.js",
                    },
                    shared: {
                        ...sharedDeps,
                        react: {
                            singleton: true,
                            requiredVersion: sharedDeps.react,
                        },
                        "react-dom": {
                            singleton: true,
                            requiredVersion: sharedDeps["react-dom"],
                        },
                    },
                }),
            ],
        }
    },
});
