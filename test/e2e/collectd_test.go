package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	apis "github.com/aneeshkp/barometer-operator/pkg/apis"
	v1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 600
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestCollectd(t *testing.T) {
	//Register with framework schema
	collectdList := &v1alpha1.CollectdList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Collectd",
			APIVersion: "collectd.barometer.com/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, collectdList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("collectd-group", func(t *testing.T) {
		t.Run("collectd", CollectdCluster)
	})
}

func CollectdCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for barometer-operator to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "barometer-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = collectdDeploymentTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
func collectdDeploymentTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create interconnect customer resource
	exampleCollectd := &v1alpha1.Collectd{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Collectd",
			APIVersion: "collectd.barometer.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-collectd",
			Namespace: namespace,
		},
		Spec: v1alpha1.CollectdSpec{
			DeploymentPlan: v1alpha1.DeploymentPlanType{
				Size:  1,
				Image: "opnfv/barometer",
			},
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleCollectd, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for example-collectd to reach 1 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-collectd", 1, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "example-collectd", Namespace: namespace}, exampleCollectd)
	if err != nil {
		return err
	}
	return nil
	//exampleCollectd.Spec.DeploymentPlan.Size = 1
	//err = f.Client.Update(goctx.TODO(), exampleCollectd)
	//if err != nil {
	//	return err
	//}

	// wait for example-interconnect to reach 4 replicas
	//return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-collectd", 4, retryInterval, timeout)
}
