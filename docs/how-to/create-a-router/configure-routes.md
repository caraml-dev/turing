# Configure routes

Routes are an essential part of your Turing Router setup. Each route is defined by its ID and the endpoint. Route can be your deployed ML model, exposed via HTTP interface, or an arbitrary non-ML web-service. Each router should be configured with at least one route and each route's ID should be unique among other routes of this router.

![](../../.gitbook/assets/routes_panel.png)

It is also required to specify, which route is default. Each router has exactly one default route and zero or more non-default routes. Depending on the configuration of Experiment Engine and Ensembler, Turing might use a response from the default route as a final response that would be sent to the calling upstream system:

| Experiment Engine | Ensembler    | Final Router Response |
| ---               | ---          | ---                   |
| Disabled          | No Ensembler | Default Route response |
| Disabled          | Custom       | Ensembler response |
| Enabled           | Standard     | Default Route response, when preferred route response cannot be obtained |
| Enabled           | Custom       | Ensembler response |

{% hint style="info" %}
You should also configure timeouts for each of the routes. The request execution will be terminated, when this timeout is exceeded during a call from Turing to the route's endpoint.
{% endhint %}
