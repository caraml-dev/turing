# End-to-end Test Infra

This directory contains utilities and configurations to intialize end-to-end
test infrastructure in Turing. The configurations for different test infra 
components can be seen from the `*.yaml` files in this directory.

In order to use the test utilities, first import the script:

```bash
source $TURING_REPO/test/e2e/setup-infra.sh
```
  

Then call the functions in the following order (ignoring certain steps e.g 
install_e2e_tools if applicable):
- install_e2e_tools
- install_kubernetes_kind_cluster
- install_local_docker_registry
- install_istio
- install_knative_serving_with_istio_controller
- install_mlp
- install_vault
- install_merlin
- build_turing_router_docker_image
- build_turing_apiserver_docker_image
- install_turing

For the function to run correctly, it is assumed the `setup-infra.sh` 
script is located in this path `$TURING_REPO/test/e2e/setup-infra.sh`