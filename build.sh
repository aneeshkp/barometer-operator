#!/bin/sh

BuildFunc(){
    operator-sdk build quay.io/aneeshkp/barometer-operator:v4.0.0
    docker push quay.io/aneeshkp/barometer-operator:v4.0.0
}

DeployFunc(){
    kubectl apply -f ./deploy/service_account.yaml

    kubectl apply -f ./deploy/role.yaml

    kubectl apply -f ./deploy/role_binding.yaml
    kubectl apply -f ./deploy/role_binding.yaml


    kubectl apply -f ./deploy/operator.yaml

    kubectl apply -f ./deploy/crds/collectd_v1alpha1_collectd_crd.yaml

    kubectl apply -f ./deploy/crds/collectd_v1alpha1_collectd_cr.yaml

}
DeployLocalFunc(){
    kubectl apply -f ./deploy/service_account.yaml
    kubectl apply -f ./deploy/role.yaml
    kubectl apply -f ./deploy/role_binding.yaml
    kubectl apply -f ./deploy/operator.yaml
    kubectl apply -f ./deploy/crds/collectd_v1alpha1_collectd_crd.yaml

}

DeleteFunc(){
    kubectl delete -f ./deploy/crds/collectd_v1alpha1_collectd_cr.yaml
    kubectl delete -f ./deploy/service_account.yaml

    kubectl delete -f ./deploy/role.yaml

    kubectl delete -f ./deploy/role_binding.yaml

    kubectl delete -f ./deploy/operator.yaml

    kubectl delete -f ./deploy/crds/collectd_v1alpha1_collectd_crd.yaml

    
                                    

}

echo "Make a choice."
    selection=

        echo "
        Operation TYPE MENU
        1 - Build 
        2 - Deploy
        3 - Deploy local
        4. Delete
    "
        echo -n "Enter selection: "
        read selection
        echo ""
        case $selection in
            1 ) BuildFunc ;;
            2 ) DeployFunc ;;
            3)  DeployLocalFunc ;;
            4) DeleteFunc ;;
            * ) echo "Please enter 1, pr 2"
        esac



#export OPERATOR_NAME=collectd-exporter operator-sdk up local --namespace=default
