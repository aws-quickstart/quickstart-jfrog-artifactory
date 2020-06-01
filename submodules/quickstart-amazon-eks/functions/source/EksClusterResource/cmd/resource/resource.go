package resource

import (
	"errors"
	"fmt"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/service/eks"
	"log"
	"runtime/debug"
)

func Create(req handler.Request, _ *Model, model *Model) (handler.ProgressEvent, error) {
	defer logPanic()
	stage := getStage(req.CallbackContext)
	switch stage {
	case InitStage:
		log.Println("Starting InitStage...")
		return createInit(req, model), nil
	case LambdaInitStage:
		log.Println("Starting InitLambdaStage...")
		return initLambda(req, model), nil
	case LambdaStablilize:
		log.Println("Starting LambdaStablilizeStage...")
		return createLambdaStabilize(req, model), nil
	case ClusterStablilize:
		log.Println("Starting ClusterStablilizeStage...")
		return createClusterStabilize(req, model), nil
	case IamAuthStage:
		log.Println("Starting IamAuthStage...")
		return createIamAuthHandler(req, model), nil
	case UpdateClusterStage:
		log.Println("Starting UpdateClusterStage...")
		return createFinalize(req, model), nil
	default:
		log.Println("Failed to identify stage.")
		return errorEvent(model, errors.New(fmt.Sprintf("Unhandled stage %s", stage))), nil
	}
}

func createInit(req handler.Request, model *Model) handler.ProgressEvent {
	eksClient := eks.New(req.Session)
	if model.Name == nil {
		model.Name = generateClusterName()
	}
	_, err := createCluster(eksClient, model, false)
	return makeEvent(model, LambdaInitStage, err)
}

func initLambda(req handler.Request, model *Model) handler.ProgressEvent {
	_, err := putFunction(req.Session, model, false)
	return makeEvent(model, LambdaStablilize, err)
}

func createLambdaStabilize(req handler.Request, model *Model) handler.ProgressEvent {
	complete, err := putFunction(req.Session, model, true)
	if complete {
		return makeEvent(model, ClusterStablilize, err)
	}
	return makeEvent(model, LambdaStablilize, err)
}

func createClusterStabilize(req handler.Request, model *Model) handler.ProgressEvent {
	eksClient := eks.New(req.Session)
	clusterComplete, err := createCluster(eksClient, model, true)
	if clusterComplete {
		return makeEvent(model, IamAuthStage, err)
	}
	return makeEvent(model, ClusterStablilize, err)
}

func createIamAuthHandler(req handler.Request, model *Model) handler.ProgressEvent {
	eksClient := eks.New(req.Session)
	err := createIamAuth(req.Session, eksClient, model)
	return makeEvent(model, UpdateClusterStage, err)
}

func createFinalize(req handler.Request, model *Model) handler.ProgressEvent {
	// Call update cluster to apply disabled public endpoint and access cidr
	eksClient := eks.New(req.Session)
	clusterComplete, err := updateCluster(eksClient, model)
	if err != nil {
		return errorEvent(model, err)
	}
	if clusterComplete {
		return makeEvent(model, CompleteStage, err)
	}
	return makeEvent(model, UpdateClusterStage, err)
}

func Read(req handler.Request, _ *Model, model *Model) (handler.ProgressEvent, error) {
	defer logPanic()
	svc := eks.New(req.Session)
	progress := readCluster(svc, model)
	return progress, nil
	//return readIamAuth(req.Session, svc, progress), nil
}

func Update(req handler.Request, _ *Model, model *Model) (handler.ProgressEvent, error) {
	defer logPanic()
	eksClient := eks.New(req.Session)
	clusterComplete, err := updateCluster(eksClient, model)
	if err != nil {
		return errorEvent(model, err), nil
	}
	functionComplete, err := putFunction(req.Session, model, false)
	if err != nil {
		return errorEvent(model, err), nil
	}
	if clusterComplete && functionComplete {
		err = updateIamAuth(req.Session, eksClient, model)
		if err != nil {
			return errorEvent(model, err), nil
		}
		return successEvent(model), nil
	}
	return inProgressEvent(model, UpdateClusterStage), nil
}

func Delete(req handler.Request, _ *Model, model *Model) (handler.ProgressEvent, error) {
	defer logPanic()
	err := deleteFunction(req.Session, model, req.CallbackContext)
	if err != nil {
		return errorEvent(model, err), nil
	}
	return deleteCluster(eks.New(req.Session), model, req.CallbackContext), nil
}

func List(req handler.Request, _ *Model, _ *Model) (handler.ProgressEvent, error) {
	defer logPanic()
	progress := listClusters(eks.New(req.Session))
	return progress, nil
}

func logPanic() {
	if r := recover(); r != nil {
		log.Println(string(debug.Stack()))
		panic(r)
	}
}
