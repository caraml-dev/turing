# Configure autoscaling

You can also configure the amount of resources (CPU/Memory) allocated for each replica of your router as well as lower and upper limits for the autoscaling.

![](../../.gitbook/assets/create_router_resources.png)

{% hint style="warning" %}
In most situations, 500m CPU and 0.5Gi memory should suffice, however, you can fine tune this configuration based on the data from Turing monitoring dashboard.
{% endhint %}
