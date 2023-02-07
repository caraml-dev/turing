#! /bin/bash

set -e
project_dir=$(git rev-parse --show-toplevel)
tee /tmp/credentials.json <<EOF
{
    "k8s_config": {
        "name": $(yq .clusters[0].name -o json "/tmp/kubeconfig"),
        "cluster": $(yq '.clusters[0].cluster' -o json "/tmp/kubeconfig"),
        "user": $(yq .users[0].user -o json "/tmp/kubeconfig")
    }
}
EOF
# get yaml file
yq e -P /tmp/credentials.json -o yaml >/tmp/temp_k8sconfig.yaml

yq '.[0].k8s_config |= load("/tmp/temp_k8sconfig.yaml").k8s_config | .[0].name |= load("'"$project_dir"'/infra/docker-compose/dev/merlin/deployment-config.yaml").[0].name' "$project_dir/api/environments-dev.yaml" >"$project_dir/api/environments-dev-w-creds.yaml"
yq '.ClusterConfig.EnsemblingServiceK8sConfig |= load("/tmp/temp_k8sconfig.yaml").k8s_config | .ClusterConfig.EnvironmentConfigPath |="./environments-dev-w-creds.yaml"' "$project_dir/api/config-dev.yaml" >"$project_dir/api/config-dev-w-creds.yaml"
