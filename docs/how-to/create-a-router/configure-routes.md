# Configure routes

Routes are an essential part of your Turing Router setup. Each route is defined by its ID and the endpoint. A route can be your deployed ML model, exposed via HTTP interface, or an arbitrary non-ML web-service. Each router should be configured with at least one route and each route's ID should be unique among other routes of this router.

The request sent to the each route is constructed from the request headers to the Router (including an
identifier `Turing-Req-Id` that is uniquely assigned to each request received by the Router), and either the request
body to the Router (if there is no Enricher) or the response body from the Enricher (if enabled).

For UPI Routers, the endpoint must support [UPI interface](https://github.com/caraml-dev/universal-prediction-interface). The default service method is `/caraml.upi.v1.UniversalPredictionService/PredictValues` and is not configurable.
![](../../.gitbook/assets/routes_panel.png)

{% hint style="info" %}
You should also configure timeouts for each of the routes. The request execution will be terminated, when this timeout is exceeded during a call from Turing to the route's endpoint.
{% endhint %}
