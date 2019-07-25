# quickstart-jfrog-artifactory

This is a Quick Start to get Enterprise Production ready Artifactory deployed into your AWS environment. 

## Deployments

The goal of this project is to have several deployment options depending on a customer's requirements.

    - New VPC deployed onto dedicated EC2 Instances.
    - Existing VPC deployed onto dedicated EC2 Instances.
    - New VPC deploying EKS, with Artifactory deployed onto the K8s cluster
    - Existing VPC deploying EKS, with Artifactory deployed onto the K8s cluster

## Development

In order to deploy a test deployment:

1. Download the repo
2. pip install the awscli (--user)
3. create a hidden folder: .ignore/
4. Inside the .ignore/ create a `params` file that is plain Text ParameterKey=DatabasePassword,ParameterValue=Password ParameterKey=KeyPairName,ParameterValue=My-SSH,ParameterKey=AvailabilityZones,ParameterValue="us-west-2a,us-west-2b"
5. Configure your `~/.aws/credentials` for use with the awscli
6. Execute the cloudformation template from inside the repo: `aws cloudformation create-stack --stack-name test --template-body file://$(pwd)/templates/jfrogartifactory-ec2-master.template --parameters $(cat .ignore/params) --capabilities CAPABILITY_IAM`

In order to test from testcat you need an override file in your home .aws directory: `~/.aws/taskcat_global_override.json`

    bash
    [  
        {
            "ParameterKey": "KeyPairName",
            "ParameterValue": "<REPLACE_ME>"
        }
    ]
Then you need to be above the repository directory and execute: `taskcat -c theflash/ci/config.yml`

### Clean up

To Delete the stack: `aws cloudformation delete-stack --stack-name test`
