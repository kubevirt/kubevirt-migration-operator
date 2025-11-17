#!/bin/bash

set -ex

source ./cluster-up/hack/common.sh
source ./cluster-up/cluster/${KUBEVIRT_PROVIDER}/provider.sh

_kubectl delete migcontroller --all --all-namespaces || echo "this is fine"
make undeploy || echo "this is fine"

port=$(./cluster-up/cli.sh ports registry | xargs)
# push to local registry provided by kvci
make docker-build IMG="localhost:${port}/kubevirt-migration-operator:latest"
make docker-push IMG="localhost:${port}/kubevirt-migration-operator:latest"
# the "cluster" (kvci VM) only understands the alias registry:5000 (which maps to localhost:${port})
make deploy IMG="registry:5000/kubevirt-migration-operator:latest"
