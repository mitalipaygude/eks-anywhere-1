package workload

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

// task related entities

type setAndValidateWorkloadTask struct{}

// SetAndValidateTask implementation

// Run createCluster performs actions needed to create the workload cluster.
func (s *setAndValidateWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
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
	return &createCluster{}
}

func (s *setAndValidateWorkloadTask) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("POC workload cluster's %s Provider setup is valid", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateCreateCluster(ctx, commandContext.ClusterSpec),
			}
		},
	}
}

func (s *setAndValidateWorkloadTask) Name() string {
	return "setup-validate"
}

func (s *setAndValidateWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *setAndValidateWorkloadTask) Checkpoint() *task.CompletedTask {
	return nil
}

type createCluster struct{}

// Run upgradeCluster performs actions needed to upgrade the management cluster.
func (s *createCluster) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating workload cluster")
	if err := commandContext.ClusterCreater.Run(ctx, commandContext.ClusterSpec, *commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	return &writeClusterConfig{}
}

func (s *createCluster) Name() string {
	return "create-workload-cluster"
}

func (s *createCluster) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *createCluster) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &writeClusterConfig{}, nil
}
