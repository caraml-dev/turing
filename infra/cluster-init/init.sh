#!/bin/bash
set -e

function parse_args {
    local args=(
        "ISTIO_BASE_HELM_VALUES_PATH"
        "ISTIO_DISCOVERY_HELM_VALUES_PATH"
        "ISTIO_INGRESS_HELM_VALUES_PATH"
    )
    for arg in ${args[@]}; do
        [[ -z "${!arg}" ]] && echo "${arg} environment variable is not set." && exit 1
    done

    echo "Args check done."
}

function print_usage {
    echo "Usage: $0 <kubeconfig file>"
    echo "  Environment Variables:"
    echo "    ISTIO_BASE_HELM_VALUES_PATH       Helm values of Istio Base."
    echo "    ISTIO_DISCOVERY_HELM_VALUES_PATH  Helm values of Istio Discovery (istiod)."
    echo "    ISTIO_INGRESS_HELM_VALUES_PATH    Helm values of Istio Ingress Gateway."
    echo
    echo "  Optional Environment Variables:"
    echo "    ISTIO_VERSION                     Istio version, default: 1.9.9."
    echo "    KNATIVE_VERSION                   Knative version, default: 0.18.3."
    echo "    KNATIVE_ISTIO_VERSION             Knative Istio version, default: 0.18.1."
}

function install_istio {
    echo "Installing Istio."

    local istio_version=${ISTIO_VERSION:-1.9.9}
    curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${istio_version} TARGET_ARCH=x86_64 sh -
    kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
    helm install istio-base "istio-${istio_version}/manifests/charts/base" \
        --namespace istio-system \
        --values ${ISTIO_BASE_HELM_VALUES_PATH} \
        --wait
    helm install istiod "istio-${istio_version}/manifests/charts/istio-control/istio-discovery" \
        --namespace istio-system \
        --values ${ISTIO_DISCOVERY_HELM_VALUES_PATH} \
        --wait
    helm install istio-ingress "istio-${istio_version}/manifests/charts/gateways/istio-ingress" \
        --namespace istio-system \
        --values ${ISTIO_INGRESS_HELM_VALUES_PATH} \
        --wait
}

function install_knative {
    echo "Installing Knative."

    kubectl apply \
        -f "https://github.com/knative/serving/releases/download/v${KNATIVE_VERSION:-0.18.3}/serving-crds.yaml"
    kubectl apply \
        -f "https://github.com/knative/serving/releases/download/v${KNATIVE_VERSION:-0.18.3}/serving-core.yaml"
    kubectl apply \
        -f "https://github.com/knative/net-istio/releases/download/v${KNATIVE_ISTIO_VERSION:-0.18.1}/net-istio.yaml"
}

parse_args $@
install_istio
install_knative

echo "Finished installing."
