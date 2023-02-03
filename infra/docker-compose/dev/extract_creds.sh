#! /bin/bash

set -e
project_dir=$(git rev-parse --show-toplevel)
infra_docker_compose_dir=$project_dir/infra/docker-compose/dev
tee credentials.json <<EOF
{
    "k8s_config": {
        "name": $(yq .clusters[0].name -o json "$infra_docker_compose_dir/kubeconfig"),
        "cluster": $(yq '.clusters[0].cluster' -o json "$infra_docker_compose_dir/kubeconfig"),
        "user": $(yq .users[0].user -o json "$infra_docker_compose_dir/kubeconfig")
    }
}
EOF
# get yaml file
yq e -P credentials.json -o yaml >temp_k8sconfig.yaml

yq '.[0].k8s_config |= load("temp_k8sconfig.yaml").k8s_config | .[0].name |= load("'"$project_dir"'/infra/docker-compose/dev/merlin/deployment-config.yaml").[0].name' "$project_dir/api/environments-dev.yaml" >"$project_dir/api/environments-dev-w-creds.yaml"
yq '.ClusterConfig.EnsemblingServiceK8sConfig |= load("temp_k8sconfig.yaml").k8s_config | .ClusterConfig.EnvironmentConfigPath |="./environments-dev-w-creds.yaml"' "$project_dir/api/config-dev.yaml" >"$project_dir/api/config-dev-w-creds.yaml"
