#/bin/bash

helm3 delete -n turing turing-test
kubectl delete serviceaccount -n turing turing-test-spark-operator

# istio
helm3 delete -n istio-system istio-base
helm3 delete -n istio-system istiod
helm3 delete -n istio-system istio-ingress

# Knative
kubectl delete ns knative-serving
kubectl delete MutatingWebhookConfiguration  -l serving.knative.dev/release=v0.18.3
kubectl delete MutatingWebhookConfiguration  -l serving.knative.dev/release=v0.18.1
kubectl delete ValidatingWebhookConfiguration  -l serving.knative.dev/release=v0.18.3
kubectl delete ValidatingWebhookConfiguration  -l serving.knative.dev/release=v0.18.1
