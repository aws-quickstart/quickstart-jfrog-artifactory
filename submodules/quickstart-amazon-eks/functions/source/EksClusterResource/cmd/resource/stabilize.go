package resource

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
)

type Stage string

const (
	InitStage          Stage = "Init"
	LambdaInitStage    Stage = "LambdaInit"
	ClusterStablilize  Stage = "ClusterStabilize"
	LambdaStablilize   Stage = "LambdaStabilize"
	IamAuthStage       Stage = "IamAuthStage"
	UpdateClusterStage Stage = "UpdateCluster"
	DeleteClusterStage Stage = "DeleteCluster"
	CompleteStage      Stage = "Complete"
)

func stabilize(svc eksiface.EKSAPI, desiredModel *Model, desiredState string) (*Model, OperationComplete, error) {
	currentModel := &Model{}
	input := &eks.DescribeClusterInput{Name: desiredModel.Name}
	response, err := svc.DescribeCluster(input)
	if err != nil {
		// if desired state is to have the cluster not found (deleted) we've succeeded
		if matchesAwsErrorCode(err, eks.ErrCodeResourceNotFoundException) && desiredState == "DELETED" {
			return nil, Complete, nil
		}
		// otherwise this is an error
		return nil, Complete, err
	}
	describeClusterToModel(*response.Cluster, currentModel)
	// status matches what we want, resource is stable
	if *response.Cluster.Status == desiredState {
		return currentModel, Complete, nil
	}
	// cluster is in a failed state
	if *response.Cluster.Status == "FAILED" {
		return currentModel, Complete, errors.New("cluster status is FAILED")
	}
	// resource is not yet stabilized
	return currentModel, InProgress, nil
}

func getStage(context map[string]interface{}) Stage {
	if context == nil {
		return InitStage
	}
	if context["Stage"] == nil {
		return InitStage
	}
	return Stage(context["Stage"].(string))
}
