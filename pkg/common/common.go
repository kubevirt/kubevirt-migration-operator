package common

// Common types and constants used by the importer and controller.
// TODO: maybe the vm cloner can use these common values

const (
	// AppKubernetesPartOfLabel is the Kubernetes recommended part-of label
	AppKubernetesPartOfLabel = "app.kubernetes.io/part-of"
	// AppKubernetesVersionLabel is the Kubernetes recommended version label
	AppKubernetesVersionLabel = "app.kubernetes.io/version"
	// AppKubernetesManagedByLabel is the Kubernetes recommended managed-by label
	AppKubernetesManagedByLabel = "app.kubernetes.io/managed-by"
	// AppKubernetesComponentLabel is the Kubernetes recommended component label
	AppKubernetesComponentLabel = "app.kubernetes.io/component"

	// The restricted SCC and particularly v2 is considered best practice for
	//  workloads that can manage without extended privileges
	RestrictedSCCName = "restricted-v2"

	// InstallerPartOfLabel provides a constant to capture our env variable "INSTALLER_PART_OF_LABEL"
	InstallerPartOfLabel = "INSTALLER_PART_OF_LABEL"
	// InstallerVersionLabel provides a constant to capture our env variable "INSTALLER_VERSION_LABEL"
	InstallerVersionLabel = "INSTALLER_VERSION_LABEL"

	// PrometheusLabelKey provides the label to indicate prometheus metrics are available in the pods.
	PrometheusLabelKey = "prometheus.migrations.kubevirt.io"
	// PrometheusServiceName is the name of the service that exposes prometheus metrics
	PrometheusServiceName = "kubevirt-migration-prometheus"
	// PrometheusLabelValue provides the label value
	PrometheusLabelValue = "true"

	// ControllerResourceName is the controller resource name
	ControllerResourceName = "kubevirt-migration-controller"
	// ControllerServiceAccountName is the name of the controller service account
	ControllerServiceAccountName = "kubevirt-migration-sa"
	// ComponentLabel is the labe applied to all non operator resources
	ComponentLabel = "migrations.kubevirt.io"
	// ConfigMapName is the name of the configmap that owns controller resources
	ConfigMapName = "kubevirt-migration-controller-config"
)
