const { ModuleFederationPlugin } = require("webpack").container;
// Remove source code dependency react-lazylog which causes problems with sharing
const { ["react-lazylog"]: undefined, ...sharedDeps } = require("./package.json").dependencies;

module.exports = ({ env }) => ({
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
