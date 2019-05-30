package collectd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	collectdv1alpha1 "github.com/aneeshkp/collectd-operator/pkg/apis/collectd/v1alpha1"
	"github.com/aneeshkp/collectd-operator/pkg/resources/deployments"
	"github.com/aneeshkp/collectd-operator/pkg/resources/serviceaccounts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

var log = logf.Log.WithName("controller_collectd")

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

	// Check if serviceaccount already exists, if not create a new one
	svcAccntFound := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, svcAccntFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new serviceaccount
		svcaccnt := serviceaccounts.NewServiceAccountForCR(instance)
		controllerutil.SetControllerReference(instance, svcaccnt, r.scheme)
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount", svcaccnt)
		err = r.client.Create(context.TODO(), svcaccnt)
		if err != nil {
			reqLogger.Error(err, "Failed to create new ServiceAccount")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get ServiceAccount")
		return reconcile.Result{}, err
	}

	configMapFound := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "collectd-config", Namespace: instance.Namespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Error(err, "ConfigMap not found... wont deploy untill config map is found\n")
		return reconcile.Result{}, err
	}
	// Set Collectd instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMapFound, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	// get the sha
	out, err := json.Marshal(configMapFound)
	if err != nil {
		return reconcile.Result{}, err
	}
	h := sha256.New()
	h.Write(out)
	currentConfigHash := fmt.Sprintf("%x", h.Sum(nil))
	reqLogger.Info("The CurrentConfig Hash : " + currentConfigHash)

	//check if deployment already exists
	//depFound := &appsv1.Deployment{}
	depFound := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, depFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := deployments.NewDaemonSetForCR(instance)
		if dep.Annotations == nil {
			dep.Annotations = make(map[string]string)
		}
		dep.Annotations["configHash"] = currentConfigHash

		// Set Collectd instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new Deployment", "Daemonset", dep)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Daemonset\n")
			return reconcile.Result{}, err
		}
		// update status
		//if reuired
		//r.client.Status().Update(context.TODO(), instance)
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Daemonset\n")
		return reconcile.Result{}, err
	}

	deployedConfigHash := depFound.Annotations["configHash"]
	//desiredConfigMap := &corev1.ConfigMap{} // where to ge desired configmap
	//eq := reflect.DeepEqual(currentConfigMap, currentConfigMap)
	reqLogger.Info("Deployed Hash : " + deployedConfigHash)

	if deployedConfigHash != currentConfigHash {
		reqLogger.Info("Change in configMap , delete deployment.")
		err = r.client.Delete(context.TODO(), depFound)
		if err != nil {
			reqLogger.Error(err, "Failed to update deployment")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil

	}

	//size := instance.Spec.DeploymentPlan.Size

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Daemonset  already exists", "DEP.Namespace", depFound.Namespace, "DEP.Name", depFound.Name)
	return reconcile.Result{}, nil
}
