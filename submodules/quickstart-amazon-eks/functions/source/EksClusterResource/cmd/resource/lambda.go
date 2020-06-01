package resource

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"io/ioutil"
	"log"
)

const (
	ZipFile            string = "k8svpc.zip"
	FunctionNamePrefix string = "k8s-api-vpc-connector-"
	Handler            string = "k8svpc"
	MemorySize         int64  = 256
	Runtime            string = "go1.x"
	Timeout            int64  = 900
)

type Event struct {
	ClusterName *string     `json:"clustername,omitempty"`
	Endpoint    *string     `json:"endpoint,omitempty"`
	CaData      []byte      `json:"cadata,omitempty"`
	AwsAuth     *IamAuthMap `json:"apiaccess,omitempty"`
	Action      Action      `json:"action,omitempty"`
}

//Status represents the status of the handler.
type Action string

const (
	CreateAction Action = "Create"
	ReadAction   Action = "Read"
	UpdateAction Action = "Update"
	DeleteAction Action = "Delete"
	ListAction   Action = "List"
)

type OperationComplete bool

const (
	Complete   OperationComplete = true
	InProgress OperationComplete = false
)

func putFunction(sess *session.Session, model *Model, reInvoke bool) (OperationComplete, error) {
	svc := lambda.New(sess)
	if reInvoke {
		return stabilizeFunction(svc, model, aws.String(FunctionNamePrefix+*model.Name))
	}
	roleArn, err := putRole(iam.New(sess))
	if err != nil {
		return Complete, err
	}

	clusterName := model.Name
	vpcConfig := *model.ResourcesVpcConfig
	err = updateFunction(svc, roleArn, clusterName, vpcConfig)
	if err != nil {
		if functionNotExists(err) {
			err = createFunction(svc, roleArn, clusterName, vpcConfig)
			if err != nil {
				return Complete, err
			}
		} else {
			return Complete, err
		}
	}
	return stabilizeFunction(svc, model, aws.String(FunctionNamePrefix+*model.Name))
}

func functionNotExists(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code() == lambda.ErrCodeResourceNotFoundException
	}
	return false
}

func createFunction(svc lambdaiface.LambdaAPI, roleArn *string, clusterName *string, vpcConfig ResourcesVpcConfig) error {
	zip, _, err := getZip()
	if err != nil {
		return err
	}
	input := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			ZipFile: zip,
		},
		FunctionName: aws.String(FunctionNamePrefix + *clusterName),
		Handler:      aws.String(Handler),
		MemorySize:   aws.Int64(MemorySize),
		Role:         roleArn,
		Runtime:      aws.String(Runtime),
		Timeout:      aws.Int64(Timeout),
		VpcConfig: &lambda.VpcConfig{
			SecurityGroupIds: aws.StringSlice(vpcConfig.SecurityGroupIds),
			SubnetIds:        aws.StringSlice(vpcConfig.SubnetIds),
		},
	}
	_, err = svc.CreateFunction(input)
	return err
}

func getZip() ([]byte, string, error) {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(ZipFile)
	hasher.Write(s)
	if err != nil {
		return nil, "", err
	}
	return s, base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

func updateFunction(svc lambdaiface.LambdaAPI, roleArn *string, clusterName *string, vpcConfig ResourcesVpcConfig) error {
	zip, hash, err := getZip()
	if err != nil {
		return err
	}
	functionOutput, err := svc.GetFunction(&lambda.GetFunctionInput{FunctionName: aws.String(FunctionNamePrefix + *clusterName)})
	if err != nil {
		return err
	}
	if hash != *functionOutput.Configuration.CodeSha256 {
		codeInput := &lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(FunctionNamePrefix + *clusterName),
			ZipFile:      zip,
		}
		_, err = svc.UpdateFunctionCode(codeInput)
		if err != nil {
			return err
		}
	}
	configInput := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(FunctionNamePrefix + *clusterName),
		Handler:      aws.String(Handler),
		MemorySize:   aws.Int64(MemorySize),
		Role:         roleArn,
		Runtime:      aws.String(Runtime),
		Timeout:      aws.Int64(Timeout),
		VpcConfig: &lambda.VpcConfig{
			SecurityGroupIds: aws.StringSlice(vpcConfig.SecurityGroupIds),
			SubnetIds:        aws.StringSlice(vpcConfig.SubnetIds),
		},
	}

	_, err = svc.UpdateFunctionConfiguration(configInput)
	return err
}

func deleteFunction(sess *session.Session, model *Model, callbackContext map[string]interface{}) error {
	if callbackContext != nil || model.Name == nil {
		return nil
	}
	svc := lambda.New(sess)
	_, err := svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(FunctionNamePrefix + *model.Name),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == lambda.ErrCodeResourceNotFoundException {
				return nil
			}
		}
	}
	return err
}

func stabilizeFunction(svc lambdaiface.LambdaAPI, model *Model, functionName *string) (OperationComplete, error) {
	for {
		output, err := svc.GetFunction(&lambda.GetFunctionInput{FunctionName: functionName})
		if err != nil {
			return Complete, err
		}
		if *output.Configuration.State == lambda.StatePending {
			return InProgress, nil
		} else if *output.Configuration.State == lambda.StateActive {
			return Complete, nil
		} else {
			errMsg := fmt.Sprintf("lambda failed to stabilize: %v[%v]: %v", *output.Configuration.State, *output.Configuration.StateReasonCode, *output.Configuration.StateReason)
			return Complete, errors.New(errMsg)
		}
	}
}

func invokeLambda(session *session.Session, svc lambdaiface.LambdaAPI, clusterName *string, iamAuthMap *IamAuthMap, action Action) (*IamAuthMap, error) {
	endpoint, caData, err := GetClusterDetails(eks.New(session), clusterName)
	if err != nil {
		return nil, err
	}
	event := Event{
		ClusterName: clusterName,
		Endpoint:    endpoint,
		CaData:      caData,
		AwsAuth:     iamAuthMap,
		Action:      action,
	}

	eventJson, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	input := &lambda.InvokeInput{
		FunctionName: aws.String(FunctionNamePrefix + *clusterName),
		Payload:      eventJson,
	}

	result, err := svc.Invoke(input)
	if err != nil {
		return nil, err
	}
	if result.FunctionError != nil {
		log.Printf("Remote execution error: %v\n", *result.FunctionError)
		errorDetails := make(map[string]string)
		err := json.Unmarshal(result.Payload, &errorDetails)
		errMsg := ""
		if err != nil {
			log.Println(err.Error())
			errMsg = fmt.Sprintf("[%v] %v", *result.FunctionError, string(result.Payload))
		} else {
			errMsg = fmt.Sprintf("[%v] %v", errorDetails["errorType"], errorDetails["errorMessage"])
		}
		return nil, errors.New(errMsg)
	}
	resp := &IamAuthMap{}
	err = json.Unmarshal(result.Payload, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
