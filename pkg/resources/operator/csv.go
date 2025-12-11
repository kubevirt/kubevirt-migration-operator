package operator

import (
	csvv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"

	corev1 "k8s.io/api/core/v1"
)

// ClusterServiceVersionData - Data arguments used to create CDI's CSV manifest
type ClusterServiceVersionData struct {
	CsvVersion         string
	ReplacesCsvVersion string
	Namespace          string
	ImagePullPolicy    string
	ImagePullSecrets   []corev1.LocalObjectReference
	IconBase64         string
	Verbosity          string

	OperatorVersion string

	ControllerImage string
	OperatorImage   string
	Rules           []rbacv1.PolicyRule
	ClusterRules    []rbacv1.PolicyRule
}

// NewClusterServiceVersion - generates CSV for CDI
func NewClusterServiceVersion(data *ClusterServiceVersionData) (*csvv1.ClusterServiceVersion, error) {
	return createClusterServiceVersion(data)
}
