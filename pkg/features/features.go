package features

// These are environment variables used as flags to enable/disable features.
const (
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	CheckpointEnabledEnvVar         = "CHECKPOINT_ENABLED"
	UseNewWorkflowsEnvVar           = "USE_NEW_WORKFLOWS"
	UseControllerForWorkloadCli     = "USE_CONTROLLER_FOR_WORKLOAD_CLI"
	ExperimentalSelfManagedClusterUpgradeEnvVar = "EXP_SELF_MANAGED_API_UPGRADE"
	ExperimentalSelfManagedClusterUpgradeGate   = "ExpSelfManagedAPIUpgrade"
)

func FeedGates(featureGates []string) {
	globalFeatures.feedGates(featureGates)
}

type Feature struct {
	Name     string
	IsActive func() bool
}

func IsActive(feature Feature) bool {
	return feature.IsActive()
}

// ClearCache is mainly used for unit tests as of now.
func ClearCache() {
	globalFeatures.clearCache()
}

func CloudStackKubeVipDisabled() Feature {
	return Feature{
		Name:     "Kube-vip support disabled in CloudStack provider",
		IsActive: globalFeatures.isActiveForEnvVar(CloudStackKubeVipDisabledEnvVar),
	}
}

func CheckpointEnabled() Feature {
	return Feature{
		Name:     "Checkpoint to rerun commands enabled",
		IsActive: globalFeatures.isActiveForEnvVar(CheckpointEnabledEnvVar),
	}
}

func UseNewWorkflows() Feature {
	return Feature{
		Name:     "Use new workflow logic for cluster management operations",
		IsActive: globalFeatures.isActiveForEnvVar(UseNewWorkflowsEnvVar),
	}
}

func UseControllerViaCLIWorkflow() Feature {
	return Feature{
		Name:     "Use new workflow logic for workload cluster creation leveraging controller via CLI",
		IsActive: globalFeatures.isActiveForEnvVar(UseControllerForWorkloadCli),
	}
}