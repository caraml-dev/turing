#!/bin/bash
set -e

function parse_args {
    local args=(
        "ISTIO_OPERATOR_PATH"
    )
    for arg in ${args[@]}; do
        [[ -z "${!arg}" ]] && echo "${arg} environment variable is not set." && exit 1
    done

    echo "Args check done."
}

function print_usage {
    echo "Usage: $0 <kubeconfig file>"
    echo "  Environment Variables:"
    echo "    ISTIO_OPERATOR_PATH    Helm values of Istio Operator."
    echo
    echo "  Optional Environment Variables:"
    echo "    ISTIO_VERSION          Istio version, default: 1.9.9."
    echo "    KNATIVE_VERSION        Knative version, default: 0.18.3."
    echo "    KNATIVE_ISTIO_VERSION  Knative Istio version, default: 0.18.1."
}

function install_istio {
    echo "Installing Istio."

    local istio_version=${ISTIO_VERSION:-1.9.9}
    curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${istio_version} TARGET_ARCH=x86_64 sh -
    kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
    ./istio-${istio_version}/bin/istioctl install -y -f ${ISTIO_OPERATOR_PATH}
    echo "Finished installing Istio."
}

function install_knative {
    echo "Installing Knative."

    kubectl apply \
        -f "https://github.com/knative/serving/releases/download/v${KNATIVE_VERSION:-0.18.3}/serving-crds.yaml"

    kubectl apply \
        -f "https://github.com/knative/serving/releases/download/v${KNATIVE_VERSION:-0.18.3}/serving-core.yaml"
    local core_apps=("activator" "autoscaler" "controller" "webhook")
    for app in ${core_apps[@]}; do
        kubectl wait -n knative-serving --for=condition=ready pod -l app=${app} --timeout=5m
    done

    kubectl apply \
        -f "https://github.com/knative/net-istio/releases/download/v${KNATIVE_ISTIO_VERSION:-0.18.1}/net-istio.yaml"
    local istio_apps=("networking-istio" "istio-webhook")
    for app in ${istio_apps[@]}; do
        kubectl wait -n knative-serving --for=condition=ready pod -l app=${app} --timeout=5m
    done

    echo "Finished installing Knative."
}

parse_args $@
install_istio
install_knative

echo "Finished installing."
