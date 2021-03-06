#!/usr/bin/env bash

# This script is a placeholder for CI until we get full OLM / CI integration

set -e
set -o pipefail

BASE_DIR=$(dirname "$0")

run () {
    echo "$@" >&2
    "$@"
}


# deploy hostpath plugin and registrar sidecar
echo "creating share CRD"
run oc apply -f ${BASE_DIR}/0000_10_projectedresource.crd.yaml
run oc apply -f ${BASE_DIR}/00-namespace.yaml
run oc apply -f ${BASE_DIR}/01-service-account.yaml
run oc apply -f ${BASE_DIR}/02-cluster-role.yaml
run oc apply -f ${BASE_DIR}/03-cluster-role-binding.yaml
run oc apply -f ${BASE_DIR}/csi-hostpath-driverinfo.yaml