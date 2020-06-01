package resource

import (
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"log"
)

func createCluster(svc eksiface.EKSAPI, model *Model, reInvoke bool) (OperationComplete, error) {
	if reInvoke {
		_, complete, err := stabilize(svc, model, "ACTIVE")
		return complete, err
	}
	input := makeCreateClusterInput(model)
	_, err := svc.CreateCluster(input)
	if err != nil {
		return Complete, err
	}
	return InProgress, nil
}

func readCluster(svc eksiface.EKSAPI, model *Model) handler.ProgressEvent {
	response, err := svc.DescribeCluster(&eks.DescribeClusterInput{Name: model.Name})
	if err != nil {
		return errorEvent(model, err)
	}
	describeClusterToModel(*response.Cluster, model)
	return successEvent(model)
}

func updateCluster(svc eksiface.EKSAPI, desiredModel *Model) (OperationComplete, error) {
	currentModel, complete, err := stabilize(svc, desiredModel, "ACTIVE")
	if err != nil {
		return Complete, err
	}
	if !complete {
		return InProgress, err
	}
	if vpcChanged(*currentModel, *desiredModel) {
		log.Println("Updating VPC config...")
		err := updateVpcConfig(svc, desiredModel)
		if err != nil {
			return Complete, err
		}
		return InProgress, nil
	}
	if loggingChanged(*currentModel, *desiredModel) {
		log.Println("Updating logging config...")
		err := updateLoggingConfig(svc, desiredModel)
		if err != nil {
			return Complete, err
		}
		return InProgress, nil
	}
	if versionChanged(*currentModel, *desiredModel) {
		log.Println("Updating kubernetes version...")
		err := updateVersionConfig(svc, desiredModel)
		if err != nil {
			return Complete, err
		}
		return InProgress, nil
	}
	return Complete, nil
}

func deleteCluster(svc eksiface.EKSAPI, model *Model, callbackContext map[string]interface{}) handler.ProgressEvent {
	if callbackContext != nil {
		_, complete, err := stabilize(svc, model, "DELETED")
		if err != nil {
			return errorEvent(model, err)
		}
		if !complete {
			return inProgressEvent(model, DeleteClusterStage)
		}
		return successEvent(model)
	}
	_, err := svc.DeleteCluster(&eks.DeleteClusterInput{Name: model.Name})
	if err != nil {
		return errorEvent(model, err)
	}
	return inProgressEvent(model, DeleteClusterStage)
}

func listClusters(svc eksiface.EKSAPI) handler.ProgressEvent {
	response, err := svc.ListClusters(&eks.ListClustersInput{})
	if err != nil {
		return errorEvent(nil, err)
	}
	models := make([]interface{}, 1)
	for _, m := range response.Clusters {
		models = append(models, &Model{Name: m})
	}
	return handler.ProgressEvent{
		ResourceModels:  models,
		OperationStatus: handler.Success,
	}
}
