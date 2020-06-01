package resource

import (
	"fmt"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/eks"
	"log"
)

const (
	callbackDelaySeconds int64 = 120
)

func errorEvent(model *Model, err error) handler.ProgressEvent {
	log.Println("Returning ERROR...")
	errorType := cloudformation.HandlerErrorCodeGeneralServiceException
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case eks.ErrCodeResourceLimitExceededException:
			errorType = cloudformation.HandlerErrorCodeServiceLimitExceeded
		case eks.ErrCodeInvalidParameterException:
			errorType = cloudformation.HandlerErrorCodeInvalidRequest
		case eks.ErrCodeUnsupportedAvailabilityZoneException:
			errorType = cloudformation.HandlerErrorCodeInvalidRequest
		case eks.ErrCodeNotFoundException:
			errorType = cloudformation.HandlerErrorCodeNotFound
		case eks.ErrCodeResourceNotFoundException:
			errorType = cloudformation.HandlerErrorCodeNotFound
		case eks.ErrCodeResourceInUseException:
			errorType = cloudformation.HandlerErrorCodeAlreadyExists
		}
	}
	return handler.ProgressEvent{
		OperationStatus:  handler.Failed,
		HandlerErrorCode: errorType,
		Message:          err.Error(),
		ResourceModel:    model,
	}
}

func successEvent(model *Model) handler.ProgressEvent {
	log.Println("Returning SUCCESS...")
	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		ResourceModel:   model,
	}
}

func inProgressEvent(model *Model, stage Stage) handler.ProgressEvent {
	log.Printf("Returning IN_PROGRESS with Id %v, next stage %v...\n", *model.Name, stage)
	return handler.ProgressEvent{
		OperationStatus:      handler.InProgress,
		ResourceModel:        model,
		Message:              fmt.Sprintf("%v in progress\n", stage),
		CallbackContext:      map[string]interface{}{"Stage": stage},
		CallbackDelaySeconds: callbackDelaySeconds,
	}
}

func makeEvent(model *Model, nextStage Stage, err error) handler.ProgressEvent {
	if err != nil {
		return errorEvent(model, err)
	}
	if nextStage == CompleteStage {
		return successEvent(model)
	}
	return inProgressEvent(model, nextStage)
}
