/*
Copyright 2025 The KubeVirt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kelseyhightower/envconfig"

	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/callbacks"
	sdkr "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/reconciler"
	migrationsv1alpha1 "kubevirt.io/kubevirt-migration-operator/api/v1alpha1"
	"kubevirt.io/kubevirt-migration-operator/pkg/common"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources/cluster"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources/namespaced"
)

var (
	log = logf.Log.WithName("migration-operator")
)

const (
	finalizerName = "operator.migrations.kubevirt.io/finalizer"

	createVersionLabel = "operator.migrations.kubevirt.io/createVersion"
	updateVersionLabel = "operator.migrations.kubevirt.io/updateVersion"
	// LastAppliedConfigAnnotation is the annotation that holds the last resource state which we put on resources under our governance
	LastAppliedConfigAnnotation = "operator.migrations.kubevirt.io/lastAppliedConfiguration"

	requeueInterval = 1 * time.Minute
)

// MigControllerReconciler reconciles a MigController object
type MigControllerReconciler struct {
	client.Client
	scheme         *runtime.Scheme
	recorder       record.EventRecorder
	namespacedArgs *namespaced.FactoryArgs
	clusterArgs    *cluster.FactoryArgs
	reconciler     *sdkr.Reconciler
	namespace      string

	getCache   func() cache.Cache
	controller controller.Controller
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) (*MigControllerReconciler, error) {
	var namespacedArgs namespaced.FactoryArgs
	namespace := GetNamespace("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	restClient := mgr.GetClient()
	clusterArgs := &cluster.FactoryArgs{
		Namespace: namespace,
		Client:    restClient,
		Logger:    log,
	}

	err := envconfig.Process("", &namespacedArgs)
	if err != nil {
		return nil, err
	}

	namespacedArgs.Namespace = namespace

	log.Info("", "VARS", fmt.Sprintf("%+v", namespacedArgs))

	scheme := mgr.GetScheme()
	uncachedClient, err := client.New(mgr.GetConfig(), client.Options{
		Scheme: scheme,
		Mapper: mgr.GetRESTMapper(),
	})
	if err != nil {
		return nil, err
	}
	recorder := mgr.GetEventRecorderFor("migcontroller-controller")

	r := &MigControllerReconciler{
		Client:         mgr.GetClient(),
		scheme:         scheme,
		recorder:       recorder,
		namespacedArgs: &namespacedArgs,
		clusterArgs:    clusterArgs,
		namespace:      namespace,
		getCache:       mgr.GetCache,
	}
	callbackDispatcher := callbacks.NewCallbackDispatcher(log, restClient, uncachedClient, scheme, namespace)
	r.reconciler = sdkr.NewReconciler(
		r, log, restClient,
		callbackDispatcher, scheme, mgr.GetCache,
		createVersionLabel, updateVersionLabel, LastAppliedConfigAnnotation,
		requeueInterval, finalizerName, true, recorder,
	).WithNamespacedCR()

	r.registerHooks()

	return r, nil
}

// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migcontrollers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migcontrollers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migcontrollers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,namespace=kubevirt-migration-system,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,namespace=kubevirt-migration-system,resources=serviceaccounts,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,namespace=kubevirt-migration-system,resources=roles;rolebindings,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions;customresourcedefinitions/status,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=list;watch;create;update;delete

// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migplans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migplans/finalizers,verbs=update
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migmigrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migmigrations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migmigrations/finalizers,verbs=update
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=migrations.kubevirt.io,resources=migclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=list;watch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=list;watch;update
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=list;watch
// +kubebuilder:rbac:groups=kubevirt.io,resources=kubevirts,verbs=list;watch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MigController object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *MigControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling MigController")

	operatorVersion := r.namespacedArgs.OperatorVersion
	cr := &migrationsv1alpha1.MigController{}
	crKey := client.ObjectKey{Namespace: req.NamespacedName.Namespace, Name: req.NamespacedName.Name}
	err := r.Client.Get(context.TODO(), crKey, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("MigController CR does not exist")
			return reconcile.Result{}, nil
		}
		log.Error(err, "Failed to get MigController object")
		return reconcile.Result{}, err
	}

	res, err := r.reconciler.Reconcile(req, operatorVersion, log)
	if err != nil {
		log.Error(err, "failed to reconcile")
	}

	return res, err
}

// createOperatorConfig creates operator config map
func (r *MigControllerReconciler) createOperatorConfig(cr client.Object) error {
	// ctrl := cr.(*migrationsv1alpha1.MigController)
	// installerLabels := util.GetRecommendedInstallerLabelsFromCr(ctrl)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.ConfigMapName,
			Namespace: r.namespace,
			Labels:    map[string]string{"operator.migrations.kubevirt.io": ""},
		},
	}
	// util.SetRecommendedLabels(cm, installerLabels, "migration-operator")

	if err := controllerutil.SetControllerReference(cr, cm, r.scheme); err != nil {
		return err
	}

	return r.Client.Create(context.TODO(), cm)
}

func (r *MigControllerReconciler) getConfigMap() (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	key := client.ObjectKey{Name: common.ConfigMapName, Namespace: r.namespace}

	if err := r.Client.Get(context.TODO(), key, cm); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return cm, nil
}

// SetController sets the controller dependency
func (r *MigControllerReconciler) SetController(controller controller.Controller) {
	r.controller = controller
	r.reconciler.WithController(controller)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MigControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new controller
	c, err := controller.New("migcontroller-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	r.SetController(c)

	if err = r.reconciler.WatchCR(); err != nil {
		return err
	}

	return nil
}

func GetNamespace(path string) string {
	if data, err := os.ReadFile(path); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "kubevirt-migration"
}
