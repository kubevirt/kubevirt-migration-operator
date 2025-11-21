package controller

import (
	"context"
	"fmt"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/callbacks"
	migrationsv1alpha1 "kubevirt.io/kubevirt-migration-operator/api/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// watch registers specific watches
func (r *MigControllerReconciler) watch() error {
	if err := r.reconciler.WatchResourceTypes(&corev1.ConfigMap{}); err != nil {
		return err
	}

	return nil
}

// preCreate creates the operator config map
func (r *MigControllerReconciler) preCreate(cr client.Object) error {
	// claim the configmap
	if err := r.createOperatorConfig(cr); err != nil {
		return err
	}
	return nil
}

// checkSanity verifies whether config map exists and is in proper relation with the cr
func (r *MigControllerReconciler) checkSanity(cr client.Object, reqLogger logr.Logger) (*reconcile.Result, error) {
	configMap, err := r.getConfigMap()
	if err != nil {
		return &reconcile.Result{}, err
	}
	if !metav1.IsControlledBy(configMap, cr) {
		ownerDeleted, err := r.configMapOwnerDeleted(configMap)
		if err != nil {
			return &reconcile.Result{}, err
		}

		if ownerDeleted || configMap.DeletionTimestamp != nil {
			reqLogger.Info("Waiting for kubevirt-migration-controller-config to be deleted before reconciling", "MigController", cr.GetName())
			return &reconcile.Result{RequeueAfter: time.Second}, nil
		}

		reqLogger.Info("Reconciling to error state, unwanted MigController object")
		result, err := r.reconciler.ReconcileError(cr, "Reconciling to error state, unwanted MigController object")
		return &result, err
	}
	return nil, nil
}

func (r *MigControllerReconciler) configMapOwnerDeleted(cm *corev1.ConfigMap) (bool, error) {
	ownerRef := metav1.GetControllerOf(cm)
	if ownerRef != nil {
		if ownerRef.Kind != "MigController" {
			return false, fmt.Errorf("unexpected configmap owner kind %q", ownerRef.Kind)
		}

		owner := &migrationsv1alpha1.MigController{}
		if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: ownerRef.Name}, owner); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		if owner.DeletionTimestamp == nil && owner.UID == ownerRef.UID {
			return false, nil
		}
	}

	return true, nil
}

func (r *MigControllerReconciler) registerHooks() {
	// Have to add these callbacks here because these are cluster scoped resources
	// and they cannot be owned by the MigController CR, since it is a namespaced resource.
	// TODO: supply fix to the operator-sdk that doesn't attempt to set owner references for cluster scoped resources
	// but instead labels them with an identifier related to the CR. Then we can check the labels here to make sure
	// we are not deleting resources that are not owned by the CR.
	r.reconciler.
		WithPreCreateHook(r.preCreate).
		WithWatchRegistrator(r.watch).
		WithSanityChecker(r.checkSanity)

	r.reconciler.AddCallback(&apiextensionsv1.CustomResourceDefinition{}, r.reconcileDeleteCRDs)
	r.reconciler.AddCallback(&rbacv1.ClusterRoleBinding{}, r.reconcileDeleteClusterRoleBinding)
	r.reconciler.AddCallback(&rbacv1.ClusterRole{}, r.reconcileDeleteClusterRole)
}

func (r *MigControllerReconciler) reconcileDeleteClusterRoleBinding(args *callbacks.ReconcileCallbackArgs) error {
	switch args.State {
	case callbacks.ReconcileStatePostDelete, callbacks.ReconcileStateOperatorDelete:
	default:
		return nil
	}

	log.Info("Deleting Cluster Role Binding", "Cluster Role Binding", args.DesiredObject.GetName())
	var clusterRoleBinding *rbacv1.ClusterRoleBinding
	if args.DesiredObject != nil {
		clusterRoleBinding = args.DesiredObject.(*rbacv1.ClusterRoleBinding)
	} else if args.CurrentObject != nil {
		clusterRoleBinding = args.CurrentObject.(*rbacv1.ClusterRoleBinding)
	} else {
		args.Logger.Info("Received cluster role binding callback with no desired/current object")
		return nil
	}

	if err := r.Client.Delete(context.TODO(), clusterRoleBinding); err != nil {
		return err
	}

	return nil
}

func (r *MigControllerReconciler) reconcileDeleteClusterRole(args *callbacks.ReconcileCallbackArgs) error {
	switch args.State {
	case callbacks.ReconcileStatePostDelete, callbacks.ReconcileStateOperatorDelete:
	default:
		return nil
	}

	log.Info("Deleting Cluster Role", "Cluster Role", args.DesiredObject.GetName())
	var clusterRole *rbacv1.ClusterRole
	if args.DesiredObject != nil {
		clusterRole = args.DesiredObject.(*rbacv1.ClusterRole)
	} else if args.CurrentObject != nil {
		clusterRole = args.CurrentObject.(*rbacv1.ClusterRole)
	} else {
		args.Logger.Info("Received cluster role callback with no desired/current object")
		return nil
	}

	if err := r.Client.Delete(context.TODO(), clusterRole); err != nil {
		return err
	}

	return nil
}

func (r *MigControllerReconciler) reconcileDeleteCRDs(args *callbacks.ReconcileCallbackArgs) error {
	switch args.State {
	case callbacks.ReconcileStatePostDelete, callbacks.ReconcileStateOperatorDelete:
	default:
		return nil
	}

	log.Info("Deleting CRD", "CRD", args.DesiredObject.GetName())
	var crd *apiextensionsv1.CustomResourceDefinition
	if args.DesiredObject != nil {
		crd = args.DesiredObject.(*apiextensionsv1.CustomResourceDefinition)
	} else if args.CurrentObject != nil {
		crd = args.CurrentObject.(*apiextensionsv1.CustomResourceDefinition)
	} else {
		args.Logger.Info("Received CRD callback with no desired/current object")
		return nil
	}

	if err := r.Client.Delete(context.TODO(), crd); err != nil {
		return err
	}

	return nil
}
