#/bin/bash

# Helper script to cleanup the infrastructure provisioned.
# Environment Variables:
#   ISTIO_VERSION         Istio Version, default 1.12.5
#   KNATIVE_VERSION       Knative version, default: 1.0.1.
#   KNATIVE_ISTIO_VERSION Knative istio version, default: 1.0.0.
#   RELEASE_NAME          Name of the helm release (Must be filled in).
#   RELEASE_NAMESPACE     Namespace of the helm release (Must be filled in).

[[ -z "${RELEASE_NAME}" ]] && echo "RELEASE_NAME environment variable is not set." && exit 1
[[ -z "${RELEASE_NAMESPACE}" ]] && echo "RELEASE_NAMESPACE environment variable is not set." && exit 1

# Istio
istio_version=${ISTIO_VERSION:-1.12.5}
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${istio_version} TARGET_ARCH=x86_64 sh -
./istio-${istio_version}/bin/istioctl x uninstall --purge -y

# Knative, this step might take some time
kubectl delete ns knative-serving

knative_versions=("${KNATIVE_VERSION:-1.0.1}" "${KNATIVE_ISTIO_VERSION:-1.0.0}")
for version in "${knative_versions[@]}"; do
    kubectl delete MutatingWebhookConfiguration -l serving.knative.dev/release="v${version}"
    kubectl delete ValidatingWebhookConfiguration -l serving.knative.dev/release="v${version}"
done

# Left over service accounts that doesn't seem to get cleaned up.
kubectl delete serviceaccount -n ${RELEASE_NAMESPACE} ${RELEASE_NAME}-spark-operator

# Left over spark operator webhook that doesn't seem to get cleaned up.
kubectl delete MutatingWebhookConfiguration -n ${RELEASE_NAMESPACE} ${RELEASE_NAME}-spark-operator-webhook-config
