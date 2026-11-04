#!/usr/bin/env bash
# Integration tests for the catalog: install Argo Workflows into the current
# kube context, check every template is accepted by the API server and passes
# `argo lint`, then run every templates/*/tests/*.yaml workflow.
#
# Usage: ARGO_VERSION=v4.1.6 hack/integration-test.sh
# Needs kubectl pointed at a disposable cluster (kind in CI).
set -euo pipefail

ARGO_VERSION=${ARGO_VERSION:-v4.1.6}
NAMESPACE=${NAMESPACE:-argo}
TEST_TIMEOUT=${TEST_TIMEOUT:-600}

cd "$(dirname "$0")/.."

bin_dir=$(mktemp -d)
trap 'rm -rf "$bin_dir"' EXIT
echo "::group::Install argo CLI ${ARGO_VERSION}"
curl -fsSL "https://github.com/argoproj/argo-workflows/releases/download/${ARGO_VERSION}/argo-linux-amd64.gz" \
  | gunzip > "$bin_dir/argo"
chmod +x "$bin_dir/argo"
argo() { "$bin_dir/argo" "$@"; }
argo version --short
echo "::endgroup::"

echo "::group::Install Argo Workflows ${ARGO_VERSION}"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
# Server-side: the full CRDs are too big for client-side apply.
kubectl apply --server-side -n "$NAMESPACE" \
  -f "https://github.com/argoproj/argo-workflows/releases/download/${ARGO_VERSION}/quick-start-minimal.yaml"
kubectl -n "$NAMESPACE" rollout status deployment/workflow-controller --timeout=300s
# Test workflows run as the namespace's default service account. The sendmail
# test creates a Secret, which the quick-start RBAC doesn't allow.
kubectl -n "$NAMESPACE" create role catalog-tests --verb=create,get --resource=secrets \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl -n "$NAMESPACE" create rolebinding catalog-tests --role=catalog-tests \
  --serviceaccount="$NAMESPACE:default" --dry-run=client -o yaml | kubectl apply -f -
echo "::endgroup::"

failures=()

echo "::group::Apply and lint templates"
for f in templates/*/manifests.yaml; do
  if ! kubectl apply --server-side -n "$NAMESPACE" -f "$f"; then
    echo "::error file=$f::rejected by the API server"
    failures+=("apply $f")
  fi
  if ! argo lint --offline "$f"; then
    echo "::error file=$f::argo lint failed"
    failures+=("lint $f")
  fi
done
echo "::endgroup::"

for f in templates/*/tests/*.yaml; do
  echo "::group::Test $f"
  wf=
  if wf=$(argo submit -n "$NAMESPACE" -o name "$f") \
    && timeout "$TEST_TIMEOUT" "$bin_dir/argo" wait -n "$NAMESPACE" "$wf" \
    && [[ $(argo get -n "$NAMESPACE" "$wf" -o json | jq -r .status.phase) == Succeeded ]]; then
    echo "[PASS] $f"
  else
    echo "::error file=$f::test workflow ${wf:-<not submitted>} did not succeed"
    [[ -n ${wf:-} ]] && { argo get -n "$NAMESPACE" "$wf" || true; argo logs -n "$NAMESPACE" "$wf" || true; }
    failures+=("test $f")
  fi
  echo "::endgroup::"
done

if ((${#failures[@]})); then
  printf 'FAILED: %s\n' "${failures[@]}"
  exit 1
fi
echo "All templates applied, linted and tested."
