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
	return &createCluster{}
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

// type CreateWorkloadViaControllerTaskPOC struct{}

// // CreateWorkloadViaControllerTaskPOC implementation

// func (s *CreateWorkloadViaControllerTaskPOC) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
// 	logger.Info("POC Creating new workload cluster using the controller")

// 	clusterName := commandContext.ClusterSpec.Cluster.Name

// 	workloadCluster := &types.Cluster{
// 		Name:               clusterName,
// 		ExistingManagement: commandContext.ManagementCluster.ExistingManagement,
// 	}

// 	workloadCluster, err := commandContext.ClusterManager.CreatePOCWorkloadCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
// 	if err != nil {
// 		commandContext.SetError(err)
// 		return &workflows.CollectDiagnosticsTask{}
// 	}
// 	commandContext.WorkloadCluster = workloadCluster

// 	return nil
// }

// func (s *CreateWorkloadViaControllerTaskPOC) Name() string {
// 	return "workload-cluster-init-poc"
// }

// func (s *CreateWorkloadViaControllerTaskPOC) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
// 	return nil, nil
// }

// func (s *CreateWorkloadViaControllerTaskPOC) Checkpoint() *task.CompletedTask {
// 	return nil
// }

// func (s *WriteWorkloadClusterConfigTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
// 	logger.Info("POC Writing cluster config file")
// 	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
// 	if err != nil {
// 		commandContext.SetError(err)
// 		return &workflows.CollectDiagnosticsTask{}
// 	}
// 	return nil
// }

// func (s *WriteWorkloadClusterConfigTask) Name() string {
// 	return "write-cluster-config"
// }

// func (s *WriteWorkloadClusterConfigTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
// 	return nil, nil
// }

// func (s *WriteWorkloadClusterConfigTask) Checkpoint() *task.CompletedTask {
// 	return nil
// }
