#!/bin/sh
operator-sdk build aneeshkp/collectd-operator:v1.0.0 

docker push aneeshkp/collectd-operator:v1.0.0

kubectl apply -f ./deploy/service_account.yaml


kubectl apply -f ./deploy/role.yaml

kubectl apply -f ./deploy/role_binding.yaml

kubectl apply -f ./deploy/operator.yaml

kubectl apply -f ./deploy/crds/collectd_v1alpha1_collectd_crd.yaml

kubectl apply -f ./deploy/crds/collectd_v1alpha1_collectd_cr.yaml


#export OPERATOR_NAME=operator-example operator-sdk up local --namespace=default