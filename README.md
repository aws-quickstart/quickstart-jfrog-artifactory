# quickstart-jfrog-artifactory

This is a Quick Start to get Enterprise Production ready Artifactory deployed into your AWS environment.

## Deployments

The goal of this project is to have several deployment options depending on a customer's requirements.

    - New VPC deployed onto dedicated EC2 Instances.
    - Existing VPC deployed onto dedicated EC2 Instances.
    - New VPC deploying EKS, with Artifactory deployed onto the K8s cluster
    - Existing VPC deploying EKS, with Artifactory deployed onto the K8s cluster

## Project Setup

    bash
    --> master template
    ----> Existing VPC
    ------> ec2-instances.

Master requires a VPC and will create, and then call Existing

Existing is then always called for this setup and will call ec2-instances

ec2-Instances builds the Primary and Secondary AutoScale/Launch Configs.

### Artifactory Configuration

Currently Artifactory is installed via Ansible utilizing roles. The main items configured are:

    bash
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
Download the submodules:

    bash
    git submodule init
    git submodule update

#### venv

    bash
    python3 -m venv ~/theflashvenv
    source ~/theflashvenv/bin/activate
    pip install awscli taskcat

#### Docker

Use the following Curl|Bash script (Feel free to look inside first) to "install" taskcat via Docker. I then moved `taskcat.docker` to `/usr/local/bin/taskcat`

    bash
    curl -s https://raw.githubusercontent.com/aws-quickstart/taskcat/master/installer/docker-installer.sh | sh
    mv taskcat.docker /usr/local/bin

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

To delete taskcat S3 buckets:
`aws s3 ls | grep taskcat | cut -d ' ' -f 3 | xargs -I {} aws s3 rb s3://{} --force`
