# Cluster Init

This is the container used to spin up Knative and Istio as they do not have publicly hosted helm charts. The `turingcluster` Helm charts will use this container to install those components. During Helm delete, the cluster will also use the `cleanup.sh` script to clean up everything that it has installed.
