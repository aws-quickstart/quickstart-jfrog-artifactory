package resource

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"strings"
)

const (
	iamRoleName      = "CloudFormation-Kubernetes-VPC"
	lambdaAssumeRole = `{
		  "Version": "2012-10-17",
		  "Statement": [
			{
			  "Effect": "Allow",
			  "Principal": {
				"Service": "lambda.amazonaws.com"
			  },
			  "Action": "sts:AssumeRole"
			}
		  ]
		}`
)

var iamPolicies = [...]string{
	"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
	"arn:aws:iam::aws:policy/service-role/AWSLambdaENIManagementAccess",
}

func getCaller(svc stsiface.STSAPI) (*string, error) {
	response, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return toRoleArn(response.Arn), nil
}

func accountIdFromArn(arn *string) *string {
	accId := strings.Split(*arn, ":")[4]
	return &accId
}

func partitionFromArn(arn *string) *string {
	partition := strings.Split(*arn, ":")[1]
	return &partition
}

func isUserArn(arn *string) bool {
	return strings.Contains(*arn, ":user/")
}

func toRoleArn(arn *string) *string {
	arnParts := strings.Split(*arn, ":")
	if arnParts[2] != "sts" || !strings.HasPrefix(arnParts[5], "assumed-role") {
		return arn
	}
	arnParts = strings.Split(*arn, "/")
	arnParts[0] = strings.Replace(arnParts[0], "assumed-role", "role", 1)
	arnParts[0] = strings.Replace(arnParts[0], ":sts:", ":iam:", 1)
	arn = aws.String(arnParts[0] + "/" + arnParts[1])
	return arn
}

func createRole(svc iamiface.IAMAPI) (*string, *string, error) {
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(lambdaAssumeRole),
		Description:              aws.String("Role used by CloudFormation to access kubernetes api's in private subnets"),
		MaxSessionDuration:       aws.Int64(3600),
		Path:                     aws.String("/"),
		RoleName:                 aws.String(iamRoleName),
	}
	output, err := svc.CreateRole(input)
	if err != nil {
		return nil, nil, err
	}
	return output.Role.RoleName, output.Role.Arn, nil
}

func attachPolicies(svc iamiface.IAMAPI, roleName *string) error {
	for _, policy := range iamPolicies {
		input := &iam.AttachRolePolicyInput{
			PolicyArn: &policy,
			RoleName:  roleName,
		}
		_, err := svc.AttachRolePolicy(input)
		if err != nil {
			return err
		}
	}
	return nil
}

func getRole(svc iamiface.IAMAPI) (*string, *string, error) {
	input := &iam.GetRoleInput{
		RoleName: aws.String(iamRoleName),
	}
	output, err := svc.GetRole(input)
	if err != nil {
		return nil, nil, err
	}
	return output.Role.RoleName, output.Role.Arn, nil
}

func putRole(svc iamiface.IAMAPI) (*string, error) {
	roleName, roleArn, err := getRole(svc)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == iam.ErrCodeNoSuchEntityException {
				roleName, roleArn, err = createRole(svc)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	err = attachPolicies(svc, roleName)
	return roleArn, err
}
