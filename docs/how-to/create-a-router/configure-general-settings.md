# Configuring General Settings

There are 3 required inputs:

**Environment**: A drop down menu of the target environment your router will be deployed to. This is set [here](https://github.com/gojek/merlin/blob/main/charts/merlin/values.yaml#L102-L130). An example used in Turing is [here](https://github.com/caraml-dev/turing/blob/main/infra/docker-compose/dev/merlin/deployment-config.yaml). As Turing manages multiple deployment environments, you are free to choose which environment your router will be deployed in.

**Name**: Name of your router deployment.

**Timeout**: Overall timeout, which when exceeded, the request execution by your Turing router will be terminated.

![](../../.gitbook/assets/general_router_settings.png)
