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
    echo "    ISTIO_OPERATOR_PATH                        Helm values of Istio Operator."
    echo
    echo "  Optional Environment Variables:"
    echo "    ISTIO_VERSION                              Istio version, default: 1.9.9."
    echo "    KNATIVE_VERSION                            Knative version, default: 0.18.3."
    echo "    KNATIVE_ISTIO_VERSION                      Knative Istio version, default: 0.18.1."
    echo "    KNATIVE_DOMAINS                            Knative domains that should be supported, comma seperated values"
    echo "    KNATIVE_REGISTRIES_SKIPPING_TAG_RESOLVING  Knative domains that should be supported, comma seperated values"
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

    # Patch knative domains
    echo "${KNATIVE_DOMAINS}" | tr ',' '\n' | while read URL; do
        if [[ ! -z ${URL} ]]; then
            local domain_template='{"data":{"{URL_TO_CHANGE}":""}}'
            local domain_patch_data=$(echo ${domain_template} | sed -e "s|{URL_TO_CHANGE}|${URL}|g")
            kubectl -n knative-serving patch configmap/config-domain --type merge -p "${domain_patch_data}"
        fi
    done

    # Patch registries
    echo "${KNATIVE_REGISTRIES_SKIPPING_TAG_RESOLVING}" | tr ',' '\n' | while read REGISTRY; do
        if [[ ! -z ${REGISTRY} ]]; then
            local registry_template='{"data":{"registriesSkippingTagResolving": "{REGISTRY_TO_CHANGE}"}}'
            local registry_patch_data=$(echo ${registry_template} | sed -e "s|{REGISTRY_TO_CHANGE}|${REGISTRY}|g")
            kubectl -n knative-serving patch configmap/config-deployment --type merge -p "${registry_patch_data}"
        fi
    done

    echo "Finished installing Knative."
}

parse_args $@
install_istio
install_knative

echo "Finished installing."
