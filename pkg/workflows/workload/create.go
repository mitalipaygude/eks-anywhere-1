package workload

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// CreateWorkload is a schema for create cluster.
type CreateWorkload struct {
	provider       providers.Provider
	clusterManager interfaces.ClusterManager
	gitOpsManager  interfaces.GitOpsManager
	writer         filewriter.FileWriter
	// eksaSpec       []byte
	eksdInstaller    interfaces.EksdInstaller
	clusterCreater   interfaces.ClusterCreater
	packageInstaller interfaces.PackageInstaller
}

// NewCreateWorkload builds a new create construct.
func NewCreateWorkload(provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	clusterCreate interfaces.ClusterCreater,
	eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageInstaller,
) *CreateWorkload {
	return &CreateWorkload{
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		clusterCreater:   clusterCreate,
		packageInstaller: packageInstaller,
	}
}

func (c *CreateWorkload) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator) error {
	logger.Info("POC New workflow creating workload cluster using the controller")
	commandContext := &task.CommandContext{
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ClusterSpec:       clusterSpec,
		Writer:            c.writer,
		Validations:       validator,
		ManagementCluster: clusterSpec.ManagementCluster,
		ClusterCreater:    c.clusterCreater,
	}

	// c.eksaSpec = eksaSpec

	err := task.NewTaskRunner(&SetAndValidateWorkloadTask{}, c.writer).RunTask(ctx, commandContext)

	return err
}
