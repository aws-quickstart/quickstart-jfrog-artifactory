package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws-quickstart/quickstart-amazon-eks-cluster-resource-provider/cmd/resource"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jinzhu/copier"
	"log"
)

func HandleRequest(_ context.Context, event resource.Event) (*resource.IamAuthMap, error) {
	eventJson, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
	}
	log.Println("event: " + string(eventJson))
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return nil, err
	}
	token, err := resource.GetToken(sess, event.ClusterName)
	if err != nil {
		return nil, err
	}
	cs, err := resource.CreateKubeClientFromToken(*event.Endpoint, *token, event.CaData)
	if err != nil {
		return nil, err
	}
	auth := &resource.IamAuthMap{}
	if event.AwsAuth != nil {
		copier.Copy(auth, event.AwsAuth)
	}
	switch event.Action {
	case resource.CreateAction:
		fmt.Println("Create event")
		err := event.AwsAuth.PushConfigMap(cs)
		if err != nil {
			return nil, err
		}
	case resource.ReadAction:
		fmt.Println("Read event")
		awsAuth, err := event.AwsAuth.GetFromCluster(cs)
		if err != nil {
			return nil, err
		}
		event.AwsAuth = awsAuth
	case resource.UpdateAction:
		fmt.Println("Update event")
		err := event.AwsAuth.PushConfigMap(cs)
		if err != nil {
			return nil, err
		}
	case resource.DeleteAction:
		fmt.Println("Delete event")
	case resource.ListAction:
		fmt.Println("List event")
	}
	return auth, nil
}

func main() {
	lambda.Start(HandleRequest)
}
