import React from "react";
import ReactDOM from "react-dom";
import "./assets/style.scss";
import * as serviceWorker from "./serviceWorker";
import App from "./App";
import * as Sentry from "@sentry/browser";

import { sentryConfig } from "./config";

Sentry.init(sentryConfig);
// Set custom tag 'app', for filtering
Sentry.setTag("app", "turing-ui");

ReactDOM.render(<App />, document.getElementById("root"));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
