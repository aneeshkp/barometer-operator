package barometer

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/aneeshkp/barometer-operator/pkg/resources/configmaps"
	"github.com/aneeshkp/barometer-operator/pkg/resources/deployments"
	"github.com/aneeshkp/barometer-operator/pkg/resources/serviceaccounts"

	collectdv1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const maxConditions = 6

var OperatorName = "UNKNOW"

var log = logf.Log.WithName("controller_barometer")

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

// Add creates a new Barometer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBarometer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("barometer-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Barometer
	err = c.Watch(&source.Kind{Type: &collectdv1alpha1.Barometer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	configFn := handler.ToRequestsFunc(func(configMapObject handler.MapObject) []reconcile.Request {
		instance := &collectdv1alpha1.Barometer{}
		//instances := &collectdv1alpha1.CollectdList{}
		depKey := types.NamespacedName{Namespace: configMapObject.Meta.GetNamespace(), Name: OperatorName}
		log.Info("OPERATOR_NAME")
		log.Info(OperatorName)
		err := mgr.GetClient().Get(context.TODO(), depKey, instance)
		if err != nil {
			log.Info("Could not find  barometer (collectd)  instance, for reconciling configmap.")
			return []reconcile.Request{}
		}
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      OperatorName,
					Namespace: configMapObject.Meta.GetNamespace(),
				}},
		}
	})
	// Watch for daemonset
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Barometer{},
	})

	// Watch for changes to secondary resource ServiceAccount and requeue the owner Interconnect
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Barometer{},
	})
	if err != nil {
		return err
	}
	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Barometer
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &collectdv1alpha1.Barometer{},
	})
	if err != nil {
		return err
	}
	//watch for secondary resources not owned by CR
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: configFn})

	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBarometer implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBarometer{}

// ReconcileBarometer reconciles a Barometer object
type ReconcileBarometer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Barometer object and makes changes based on the state read
// and what is in the Barometer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBarometer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Barometer")

	// Fetch the Barometer instance
	instance := &collectdv1alpha1.Barometer{}
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

	OperatorName = instance.GetName()
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

	returnValues = r.CheckForConfigMap(instance, reqLogger)
	if returnValues.err != nil {
		return reconcile.Result{}, err
	} else if returnValues.reQueue {
		return reconcile.Result{Requeue: true}, nil
	}

	returnValues = r.ReconcileDeployment(instance, returnValues.hash256String, reqLogger)
	if returnValues.err != nil {
		return reconcile.Result{}, err
	} else if returnValues.reQueue {
		return reconcile.Result{Requeue: true}, nil
	}

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
func (r *ReconcileBarometer) UpdateCondition(instance *collectdv1alpha1.Barometer, reason string, reqLogger logr.Logger) error {
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
func (r *ReconcileBarometer) ReconcileServiceAccount(instance *collectdv1alpha1.Barometer, reqLogger logr.Logger) ReturnValues {
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

//CheckForConfigMap  ..If configmap doesn't exist , do not deploy/
func (r *ReconcileBarometer) CheckForConfigMap(instance *collectdv1alpha1.Barometer, reqLogger logr.Logger) ReturnValues {
	configMapFound := &corev1.ConfigMap{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.DeploymentPlan.ConfigName, Namespace: instance.Namespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("ConfigMap not found...  deployment will continue with default configurations. ")
		return ReturnValues{"", false, nil}
	} else if err != nil {
		reqLogger.Error(err, "Error loading  configmap\n")
		return ReturnValues{"", false, err}
	}

	// get the sha
	/*out, err := json.Marshal(configMapFound.Data)
	if err != nil {
		return ReturnValues{"", false, err}
	}
	h := sha256.New()
	_, err = h.Write(out)
	if err != nil {
		reqLogger.Info("ERROR reading config ")
		return ReturnValues{"", false, err}
	}
	currentConfigHash := fmt.Sprintf("%x", h.Sum(nil))
	*/
	// configMap resource version will be an env in Che deployment to easily update it when a ConfigMap changes
	// which will automatically trigger Che rolling update
	currentConfigVersion := configMapFound.ResourceVersion

	reqLogger.Info("Configmap exists", "Configmap.Namespace", configMapFound.Namespace, "ConfigMap.Name", configMapFound.Name)

	return ReturnValues{currentConfigVersion, false, nil}
}

//ReconcileConfigMap  ..DON'T use this function for now... /
func (r *ReconcileBarometer) ReconcileConfigMap(instance *collectdv1alpha1.Barometer, reqLogger logr.Logger) ReturnValues {
	configMapFound := &corev1.ConfigMap{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.DeploymentPlan.ConfigName, Namespace: instance.Namespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("ConfigMap not found...  default config issued  :" + instance.Spec.DeploymentPlan.ConfigName + "\n")
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
	out, err := json.Marshal(configMapFound.Data)
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
func (r *ReconcileBarometer) ReconcileDeployment(instance *collectdv1alpha1.Barometer, cmVersion string, reqLogger logr.Logger) ReturnValues {

	//check if deployment already exists
	depFound := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, depFound)
	dep := &appsv1.DaemonSet{}

	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		if cmVersion == "" {
			dep = deployments.NewDefaultDaemonSetForCR(instance)
			reqLogger.Info("Creating a new Deployment with default configurations (use configmap to modify Barometer conf)", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		} else {
			dep = deployments.NewDaemonSetForCR(instance, cmVersion)
			reqLogger.Info("Creating a new Deployment with configMap", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		}

		if err := controllerutil.SetControllerReference(instance, dep, r.scheme); err != nil {
			return ReturnValues{"", false, err}
		}

		// Define a new deployment
		if dep.Annotations == nil {
			dep.Annotations = make(map[string]string)
		}
		dep.Annotations["configversion"] = cmVersion
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Error creating deployment for Deployment.Namespace"+depFound.Namespace+"Deployment.Name"+depFound.Name)
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
	deployedCmVersion := depFound.Annotations["configversion"]
	reqLogger.Info("CurrentConfigVersion: " + cmVersion)
	reqLogger.Info("deployedConfigVersion: " + deployedCmVersion)

	if cmVersion != deployedCmVersion {
		r.UpdateCondition(instance, "Configuration changed", reqLogger)
		reqLogger.Info("Change in configMap , delete deployment.")
		err = r.client.Delete(context.TODO(), depFound)

		if err != nil {
			reqLogger.Error(err, "Failed to delete deployment")
			return ReturnValues{"", false, err}
		}
		r.UpdateCondition(instance, "Deleteing daemonset for config updates.", reqLogger)
		return ReturnValues{"", true, nil}
	}
	return ReturnValues{"", false, nil}
}
