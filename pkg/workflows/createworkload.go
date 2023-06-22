package workflows

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/types"
	// "github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"

	// "github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type CreateWorkload struct {
	provider       providers.Provider
	clusterManager interfaces.ClusterManager
	gitOpsManager  interfaces.GitOpsManager
	writer         filewriter.FileWriter
	eksaSpec       []byte
}

func NewCreateWorkload(provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
) *CreateWorkload {
	return &CreateWorkload{
		provider:       provider,
		clusterManager: clusterManager,
		gitOpsManager:  gitOpsManager,
		writer:         writer,
	}
}

func (c *CreateWorkload) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator, eksaSpec []byte) error {
	logger.Info("POC New workflow creating workload cluster using the controller")
	commandContext := &task.CommandContext{
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ClusterSpec:       clusterSpec,
		Writer:            c.writer,
		Validations:       validator,
		ManagementCluster: clusterSpec.ManagementCluster,
	}

	c.eksaSpec = eksaSpec

	err := task.NewTaskRunner(&SetAndValidateWorkloadTask{}, c.writer).RunTask(ctx, commandContext)

	return err
}

// task related entities

type SetAndValidateWorkloadTask struct{}

type WriteWorkloadClusterConfigTask struct{}

// type CreateWorkloadClusterTask struct{}

// SetAndValidateTask implementation

func (s *SetAndValidateWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("POC Performing setup and validations")
	runner := validations.NewRunner()
	runner.Register(s.providerValidation(ctx, commandContext)...)
	runner.Register(commandContext.GitOpsManager.Validations(ctx, commandContext.ClusterSpec)...)
	runner.Register(commandContext.Validations.PreflightValidations(ctx)...)

	err := runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &CreateWorkloadViaControllerTaskPOC{}
}

func (s *SetAndValidateWorkloadTask) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("POC %s Provider setup is valid", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateCreateCluster(ctx, commandContext.ClusterSpec),
			}
		},
	}
}

func (s *SetAndValidateWorkloadTask) Name() string {
	return "setup-validate"
}

func (s *SetAndValidateWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *SetAndValidateWorkloadTask) Checkpoint() *task.CompletedTask {
	return nil
}

type CreateWorkloadViaControllerTaskPOC struct{}

// CreateWorkloadViaControllerTaskPOC implementation

func (s *CreateWorkloadViaControllerTaskPOC) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("POC Creating new workload cluster using the controller")

	clusterName := commandContext.ClusterSpec.Cluster.Name

	workloadCluster := &types.Cluster{
		Name:               clusterName,
		ExistingManagement: commandContext.ManagementCluster.ExistingManagement,
	}

	workloadCluster, err := commandContext.ClusterManager.CreatePOCWorkloadCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.WorkloadCluster = workloadCluster

	return nil
}

func (s *CreateWorkloadViaControllerTaskPOC) Name() string {
	return "workload-cluster-init-poc"
}

func (s *CreateWorkloadViaControllerTaskPOC) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *CreateWorkloadViaControllerTaskPOC) Checkpoint() *task.CompletedTask {
	return nil
}

func (s *WriteWorkloadClusterConfigTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("POC Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return nil
}

func (s *WriteWorkloadClusterConfigTask) Name() string {
	return "write-cluster-config"
}

func (s *WriteWorkloadClusterConfigTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *WriteWorkloadClusterConfigTask) Checkpoint() *task.CompletedTask {
	return nil
}
