# quickstart-jfrog-artifactory

This is a Quick Start to get Enterprise Production ready Artifactory deployed into your AWS environment. 

## Deployments

The goal of this project is to have several deployment options depending on a customer's requirements.

    - New VPC deployed onto dedicated EC2 Instances.
    - Existing VPC deployed onto dedicated EC2 Instances.
    - New VPC deploying EKS, with Artifactory deployed onto the K8s cluster
    - Existing VPC deploying EKS, with Artifactory deployed onto the K8s cluster

## Deployment from Command line

In order to deploy a test deployment:

1. Download the repo
2. pip install the awscli (--user)
3. create a hidden folder: .ignore/
4. Inside the .ignore/ create a `params` file that is plain Text ParameterKey=DatabasePassword,ParameterValue=Password ParameterKey=KeyPairName,ParameterValue=My-SSH,ParameterKey=AvailabilityZones,ParameterValue="us-west-2a,us-west-2b"
5. Configure your `~/.aws/credentials` for use with the awscli
6. Execute the cloudformation template from inside the repo: `aws cloudformation create-stack --stack-name test --template-body file://$(pwd)/templates/jfrog-artifactory-ec2-master.template --parameters $(cat .ignore/params) --capabilities CAPABILITY_NAMED_IAM`

## Testing with TaskCat

### Pre-Reqs

To install [taskcat](#https://aws-quickstart.github.io/install-taskcat.html)

#### venv

    bash
    python3 -m venv ~/theflashvenv
    source ~/theflashvenv/bin/activate
    pip install awscli taskcat

#### Docker

@chris to fill out :)

### Testing

In order to test from taskcat you need an override file in your home .aws directory: `~/.aws/taskcat_global_override.json`

    bash
    [  
        {
            "ParameterKey": "KeyPairName",
            "ParameterValue": "<REPLACE_ME>"
        }
    ]

Please also verify the `ci/config.yml` is updated with the region you wish to deploy to. The rest of the parameters should be answered in the `ci/<test>.json`

Then you need to be above the repository directory and execute: `taskcat -c theflash/ci/config.yml`.

### Clean up

To Delete the stack: `aws cloudformation delete-stack --stack-name test`
