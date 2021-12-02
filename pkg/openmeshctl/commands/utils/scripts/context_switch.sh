#!/bin/bash -ex

kubeContext=$0

# set current context to the given kubeContext
kubectl config use-context ${kubeContext}
