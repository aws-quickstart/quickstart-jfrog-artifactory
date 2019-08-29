# quickstart-jfrog-artifactory

This is a Quick Start to get Enterprise Production ready Artifactory deployed into your AWS environment.

## Deployments

The goal of this project is to have several deployment options depending on a customer's requirements.

    - New/Existing VPC deployed onto dedicated EC2 Instances.
    - New/Existing VPC deploying EKS, with Artifactory deployed onto the K8s cluster
    - New/Existing VPC deploying ECS, with Artifactory deployed as an ECS Service

## Project Setup

    --> master template
    ----> Existing VPC
    ------> {Deployment Type}
    --------> Core-Infrastructure

Master creates a new VPC, and then call Existing `Deployment Type` stack.

Existing `Deployment Type` is then always call the required nested stacks for that deployment. All stacks have a dependency on the `jfrog-artifactory-core-infrastucture` which configures the S3 bucket and RDS database for the deployment.

### Artifactory Configuration

Currently Artifactory can be deployed via EC2, ECS, and EKS. For the EC2 and ECS versions, Artifactory is installed via Ansible utilizing roles. When using EKS it is deployed using their [helm charts](https://github.com/jfrog/charts).

#### Ansible Role configuration

The role's structure is per the below tree:

    artifactory
     ├── README.md
     ├── defaults
     │   └── main.yml
     ├── files
     │   ├── inactiveServerCleaner.groovy
     │   ├── installer-info.json
     │   └── nginx.conf
     ├── handlers
     │   └── main.yml
     ├── meta
     │   └── main.yml
     ├── tasks
     │   ├── configure.yml
     |   ├── configure_ecs.yml
     │   ├── install.yml
     │   ├── main.yml
     │   └── nginx-setup.yml
     └── templates
         ├── artifactory.cluster.license.j2
         ├── artifactory.conf.j2
         ├── binarystore.xml.j2
         ├── certificate.key.j2
         ├── certificate.pem.j2
         ├── db.properties.j2
         ├── ha-node.properties.j2
         └── master.key.j2

The Templates are per documentation. For the ha-node(port set to 0) please see this [link](https://jfrog.com/knowledge-base/why-the-membership-port-in-the-ha-configuration-is-set-to-0/)

`configure_ecs.yml` is a special task for configuring docker hosts. It follows the structure from the official Artifactory [docker compose](https://github.com/jfrog/artifactory-docker-examples/tree/master/docker-compose/artifactory)

## Deployment from Command line

In order to deploy a test deployment:

1. Download the repo
2. `git submodule init; git submodule update` inside the repo.
3. pip install the awscli (--user)
4. create a hidden folder: .ignore/
5. Inside the .ignore/ create a `params` file that is plain Text ParameterKey=DatabasePassword,ParameterValue=Password ParameterKey=KeyPairName,ParameterValue=My-SSH,ParameterKey=AvailabilityZones,ParameterValue="us-west-2a,us-west-2b"
6. Configure your `~/.aws/credentials` for use with the awscli
7. Execute the cloudformation template from inside the repo: `aws cloudformation create-stack --stack-name test --template-body file://$(pwd)/templates/jfrog-artifactory-ec2-master.template --parameters $(cat .ignore/params) --capabilities CAPABILITY_NAMED_IAM`

## Testing with TaskCat

### Pre-Reqs

To install [taskcat](#https://aws-quickstart.github.io/install-taskcat.html)
Download the submodules:

    git submodule init
    git submodule update

NOTE: if you are building the EKS version of this deployment you will need to do the same commands from within the quickstart-amazon-eks (At least verify git updated the submodules).

#### venv

    python3 -m venv ~/cloudformationvenv
    source ~/cloudformationvenv/bin/activate
    pip install awscli taskcat

#### Docker

Use the following Curl|Bash script (Feel free to look inside first) to "install" taskcat via Docker. I then moved `taskcat.docker` to `/usr/local/bin/taskcat`

    curl -s https://raw.githubusercontent.com/aws-quickstart/taskcat/master/installer/docker-installer.sh | sh
    mv taskcat.docker /usr/local/bin

### Testing

In order to test from taskcat you need an override file in your home .aws directory: `~/.aws/taskcat_global_override.json`

    [  
        {
            "ParameterKey": "KeyPairName",
            "ParameterValue": "<REPLACE_ME>"
        }
    ]

Please also verify the `ci/config.yml` is updated with the region you wish to deploy to. The rest of the parameters should be answered in the `ci/<test>.json` : `jfrog-artifactory-new-vpc-ec2.json`

NOTE: We have seen issues running taskcat under the following conditions, please verify:
    * Your Environment variables for AWS are what you want as they override your `~/.aws/credentials` and `~/.aws/config`
    * You have initialized and updated the git submodules
    * You Account has the correct IAM Permissions to execute in the region.
    * Your default region and test region match.

Then you need to be above the repository directory and execute: `taskcat -c quickstart-jfrog-artifactory/ci/config.yml`.

### Clean up

To Delete the stack: `aws cloudformation delete-stack --stack-name test`

To delete taskcat S3 buckets:
`aws s3 ls | grep taskcat | cut -d ' ' -f 3 | xargs -I {} aws s3 rb s3://{} --force`
