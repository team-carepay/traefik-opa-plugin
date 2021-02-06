#!/usr/bin/env bash
set -euf -o pipefail

kind delete clusters kind || echo "kind was not started"
kind create cluster --config=kind/kind.yaml --wait 200s
kubectl cluster-info --context kind-kind
if helm repo list | grep traefik; then
  echo "helm repo alread present"
else
  helm repo add traefik https://helm.traefik.io/traefik
  helm repo update
fi
helm install --values traefik/values.yaml traefik traefik/traefik
kubectl apply -k traefik
kubectl apply -k opa
kubectl apply -k example-app
