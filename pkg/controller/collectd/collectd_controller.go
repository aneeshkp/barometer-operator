package collectd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	collectdv1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
	"github.com/aneeshkp/barometer-operator/pkg/resources/configmaps"
	"github.com/aneeshkp/barometer-operator/pkg/resources/deployments"
	"github.com/aneeshkp/barometer-operator/pkg/resources/serviceaccounts"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const maxConditions = 6

var log = logf.Log.WithName("controller_collectd")

//ReturnValues ...
type ReturnValues struct {
	hash256String string
	reQueue       bool
	err           error
}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Collectd Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconciappsle.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCollectd{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("collectd-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Collectd
	err = c.Watch(&source.Kind{Type: &collectdv1alpha1.Collectd{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for configmap
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Collectd{},
	})

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Collectd
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &collectdv1alpha1.Collectd{},
	//})
	//if err != nil {
	//	return err
	//}

	// Watch for daemonset
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Collectd{},
	})

	// Watch for changes to secondary resource ServiceAccount and requeue the owner Interconnect
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Collectd{},
	})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Collectd
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Collectd{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCollectd{}

// ReconcileCollectd reconciles a Collectd object
type ReconcileCollectd struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Collectd object and makes changes based on the state read
// and what is in the Collectd.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCollectd) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Collectd")

	// Fetch the Collectd instance
	instance := &collectdv1alpha1.Collectd{}

	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reqLogger.Info("CONFIG---------------" + instance.Spec.DeploymentPlan.ConfigName)
	// Assign the generated resource version to the status
	if instance.Status.RevNumber == "" {
		instance.Status.RevNumber = instance.ObjectMeta.ResourceVersion
		r.UpdateCondition(instance, "provision spec to desired state", reqLogger)
	}

	// Check if serviceaccount already exists, if not create a new one
	returnValues := r.ReconcileServiceAccount(instance, reqLogger)
	if returnValues.err != nil {
		return reconcile.Result{}, err
	} else if returnValues.reQueue {
		return reconcile.Result{Requeue: true}, nil
	}

	returnValues = r.ReconcileConfigMap(instance, reqLogger)
	if returnValues.err != nil {
		return reconcile.Result{}, err
	} else if returnValues.reQueue {
		return reconcile.Result{Requeue: true}, nil
	}

	//desiredConfigMap := &corev1.ConfigMap{} // where to ge desired configmap
	//eq := reflect.DeepEqual(currentConfigMap, currentConfigMap)
	returnValues = r.ReconcileDeployment(instance, returnValues.hash256String, reqLogger)
	if returnValues.err != nil {
		return reconcile.Result{}, err
	} else if returnValues.reQueue {
		return reconcile.Result{Requeue: true}, nil
	}

	//size := instance.Spec.DeploymentPlan.Size

	// Pod already exists - don't requeue

	return reconcile.Result{}, nil
}

func addCondition(conditions []collectdv1alpha1.CollectdCondition, condition collectdv1alpha1.CollectdCondition) []collectdv1alpha1.CollectdCondition {
	size := len(conditions) + 1
	first := 0
	if size > maxConditions {
		first = size - maxConditions
	}
	return append(conditions, condition)[first:size]
}

//UpdateCondition ...
func (r *ReconcileCollectd) UpdateCondition(instance *collectdv1alpha1.Collectd, reason string, reqLogger logr.Logger) error {
	// update status
	// update status
	condition := collectdv1alpha1.CollectdCondition{
		Type:           collectdv1alpha1.CollectdConditionProvisioning,
		Reason:         reason,
		TransitionTime: metav1.Now(),
	}
	instance.Status.Conditions = addCondition(instance.Status.Conditions, condition)
	r.client.Status().Update(context.TODO(), instance)
	return nil
}

//ReconcileServiceAccount  ...
func (r *ReconcileCollectd) ReconcileServiceAccount(instance *collectdv1alpha1.Collectd, reqLogger logr.Logger) ReturnValues {
	svcaccnt := serviceaccounts.NewServiceAccountForCR(instance)

	// Set OutgoingPortal instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, svcaccnt, r.scheme); err != nil {
		return ReturnValues{"", false, err}
	}
	svcAccntFound := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: svcaccnt.Name, Namespace: svcaccnt.Namespace}, svcAccntFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "svcaccnt.Namespace", svcaccnt.Namespace, "svcaccnt.Name", svcaccnt.Name)
		err = r.client.Create(context.TODO(), svcaccnt)
		if err != nil {
			reqLogger.Error(err, "Failed to create new ServiceAccount")
			return ReturnValues{"", false, err}
		}
		return ReturnValues{"", true, err}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ServiceAccount")
		return ReturnValues{"", false, err}
	}
	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: SvcAccnt already exists", "svcaccnt.Namespace",
		svcAccntFound.Namespace, "svcaccnt.Name", svcAccntFound.Name)
	return ReturnValues{"", false, nil}
}

//ReconcileConfigMap  ../
func (r *ReconcileCollectd) ReconcileConfigMap(instance *collectdv1alpha1.Collectd, reqLogger logr.Logger) ReturnValues {
	configMapFound := &corev1.ConfigMap{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.DeploymentPlan.ConfigName, Namespace: instance.Namespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("ConfigMap not found... Creating default configmap :" + instance.Spec.DeploymentPlan.ConfigName + "\n")
		configmap := configmaps.NewConfigMapForCR(instance)
		if err := controllerutil.SetControllerReference(instance, configmap, r.scheme); err != nil {
			reqLogger.Info("ERROR createing owner to config")
			return ReturnValues{"", false, err}
		}
		err = r.client.Create(context.TODO(), configmap)
		if err != nil {
			r.UpdateCondition(instance, "Error creating default Configuration", reqLogger)
			reqLogger.Error(err, "ERROR createing config\n")
			return ReturnValues{"", false, err}
		}

		return ReturnValues{"", true, nil}
	} else if err != nil {
		reqLogger.Error(err, "Error loading  configmap\n")
		return ReturnValues{"", false, err}
	}

	// get the sha
	out, err := json.Marshal(configMapFound)
	if err != nil {
		return ReturnValues{"", false, err}
	}
	h := sha256.New()
	_, err = h.Write(out)
	if err != nil {
		reqLogger.Info("ERROR reading config hah")
		return ReturnValues{"", false, err}
	}
	currentConfigHash := fmt.Sprintf("%x", h.Sum(nil))
	reqLogger.Info("Skip reconcile: Configmap already exists", "Configmap.Namespace", configMapFound.Namespace, "ConfigMap.Name", configMapFound.Name)

	return ReturnValues{currentConfigHash, false, nil}
}

//ReconcileDeployment  ...
func (r *ReconcileCollectd) ReconcileDeployment(instance *collectdv1alpha1.Collectd, currentConfigHash string, reqLogger logr.Logger) ReturnValues {
	//check if deployment already exists
	depFound := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := deployments.NewDaemonSetForCR(instance)
		if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
			return ReturnValues{"", false, err}
		}
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		r.UpdateCondition(instance, "Created Default Configuration", reqLogger)

		// Define a new deployment
		if dep.Annotations == nil {
			dep.Annotations = make(map[string]string)
		}
		dep.Annotations["configHash"] = currentConfigHash
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return ReturnValues{"", false, err}
		}
		r.UpdateCondition(instance, "Creating new deployment", reqLogger)
		return ReturnValues{"", true, nil}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Daemonset\n")
		return ReturnValues{"", false, err}
	}
	// Deployment already exists - don't requeue
	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", depFound.Namespace, "Deployment.Name", depFound.Name)
	deployedConfigHash := depFound.Annotations["configHash"]
	if deployedConfigHash != currentConfigHash {
		r.UpdateCondition(instance, "Configuration changed", reqLogger)
		reqLogger.Info("Change in configMap , delete deployment.")
		err = r.client.Delete(context.TODO(), depFound)
		r.UpdateCondition(instance, "Deleteing daemonset for config updates", reqLogger)
		if err != nil {
			reqLogger.Error(err, "Failed to update deployment")
			return ReturnValues{"", false, err}
		}
		return ReturnValues{"", true, nil}
	}
	return ReturnValues{"", false, nil}
}
