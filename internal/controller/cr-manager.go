package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"

	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk"

	migrationsv1alpha1 "kubevirt.io/kubevirt-migration-operator/api/v1alpha1"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources/cluster"
	"kubevirt.io/kubevirt-migration-operator/pkg/resources/namespaced"
)

// Status provides migcontroller status sub-resource
func (r *MigControllerReconciler) Status(cr client.Object) *sdkapi.Status {
	return &cr.(*migrationsv1alpha1.MigController).Status.Status
}

// Create creates new migcontroller resource
func (r *MigControllerReconciler) Create() client.Object {
	return &migrationsv1alpha1.MigController{}
}

// GetDependantResourcesListObjects provides slice of List resources corresponding to migration-controller-dependant resource types
func (r *MigControllerReconciler) GetDependantResourcesListObjects() []client.ObjectList {
	return []client.ObjectList{
		&extv1.CustomResourceDefinitionList{},
		&rbacv1.ClusterRoleBindingList{},
		&rbacv1.ClusterRoleList{},
		&appsv1.DeploymentList{},
		&corev1.ServiceList{},
		&rbacv1.RoleBindingList{},
		&rbacv1.RoleList{},
		&corev1.ServiceAccountList{},
	}
}

// IsCreating checks whether operator config is missing (which means it is create-type reconciliation)
func (r *MigControllerReconciler) IsCreating(_ client.Object) (bool, error) {
	configMap, err := r.getConfigMap()
	if err != nil {
		return true, nil
	}
	return configMap == nil, nil
}

func (r *MigControllerReconciler) getNamespacedArgs(cr *migrationsv1alpha1.MigController) *namespaced.FactoryArgs {
	result := *r.namespacedArgs

	if cr != nil {
		if cr.Spec.ImagePullPolicy != "" {
			result.PullPolicy = string(cr.Spec.ImagePullPolicy)
		}
		if cr.Spec.PriorityClass != nil && string(*cr.Spec.PriorityClass) != "" {
			result.PriorityClassName = string(*cr.Spec.PriorityClass)
		}
		// Verify the priority class name exists.
		priorityClass := &schedulingv1.PriorityClass{}
		if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: result.PriorityClassName}, priorityClass); err != nil {
			// Any error we cannot determine if priority class exists.
			result.PriorityClassName = ""
		}
		result.InfraNodePlacement = &cr.Spec.Infra
	}

	return &result
}

// GetAllResources provides slice of resources the migration controller depends on
func (r *MigControllerReconciler) GetAllResources(crObject client.Object) ([]client.Object, error) {
	cr := crObject.(*migrationsv1alpha1.MigController)
	var resources []client.Object

	if true {
		crs, err := cluster.CreateAllStaticResources(r.clusterArgs)
		if err != nil {
			sdk.MarkCrFailedHealing(cr, r.Status(cr), "CreateResources", "Unable to create all resources", r.recorder)
			return nil, err
		}

		resources = append(resources, crs...)
	}

	nsrs, err := namespaced.CreateAllResources(r.getNamespacedArgs(cr))
	if err != nil {
		sdk.MarkCrFailedHealing(cr, r.Status(cr), "CreateNamespaceResources", "Unable to create all namespaced resources", r.recorder)
		return nil, err
	}

	resources = append(resources, nsrs...)

	// drs, err := cluster.CreateAllDynamicResources(r.clusterArgs)
	// if err != nil {
	// 	sdk.MarkCrFailedHealing(cr, r.Status(cr), "CreateDynamicResources", "Unable to create all dynamic resources", r.recorder)
	// 	return nil, err
	// }

	// resources = append(resources, drs...)

	return resources, nil
}
