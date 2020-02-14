# quickstart-jfrog-artifactory-Trace3-Internal

This is a Quick Start to get Enterprise Production ready Artifactory deployed into your AWS environment. There are several parts to a Quickstart.

## Getting started

To complete a Quickstart you need the following items:

* Quickstart Word document (QuickStart Guide)
* Diagrams
* CloudFormation code
* Taskcat output of a successful deployment.

## Quickstart Word document

For this specific quickstart all documentation is being stored in [Box](https://trace3.box.com/s/e0infq2amuefkxqllbveq60xorpeee7f)
Note the Word document has the diagrams within it.

## Diagrams

Diagrams for this specific project are embedded within the project and also part of this [powerpoint](https://trace3.box.com/s/m97mbf4yfazdm8ruhmfsdd9x4lycukg6)
You will need to ensure you download the latest templates for the diagrams. Note: When using icons for services such as ECS, RDS, etc. they need to be within the AWS cloud box.

## CloudFormation code

Ensure Parameters fit the diction that is required. No Camel Case for them, need to end with a period and be spelled out where possible. This can take a considerable amount of effort during the approval page if you do not follow the proper way. DO not copy other CF templates as they are wrong, the templates within this should conform.

## Taskcat

Taskcat information can be found in the README.md.
We have extra YAML configured test examples within our internal repo. If you are debugging a stack, such as EKS you can deploy the whole stack, then just delete the "core" stack. Remeber, the name of the template will be somewhere in the name of the stack. Doing the core-workload will delete the helm deployment. You can then use the jfrog-artifactory-eks-core test example and fill in your variables. Use the Parameters of the core as a great starting point, as well as, the Outputs of the other stacks like core-infrastructure for the S3 key and properties. Then running taskcat with that file enabled will allow you to iterate much quicker than an entire stack creation/deletion.

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

    python3 -m venv venv
    source venv/bin/activate
    pip install awscli taskcat

#### Docker

Use the following Curl|Bash script (Feel free to look inside first) to "install" taskcat via Docker. I then moved `taskcat.docker` to `/usr/local/bin/taskcat`

    curl -s https://raw.githubusercontent.com/aws-quickstart/taskcat/master/installer/docker-installer.sh | sh
    mv taskcat.docker /usr/local/bin

### Testing

Please also verify the `.taskcat.yml` is updated with the region you wish to deploy to. The rest of the parameters should be answered in the `.taskcat.yml` for variables needed. Then a `.taskcat_overrides.yml` needs to be created to override any parameters with "override" for their value. as follows:

    bash
    RemoteAccessCIDR: `curl -v4 ifconfig.io`
    AccessCIDR: `curl -v4 ifconfig.io`
    KeyPairName: `your-keypair`
    MasterKey: 1ce2be4490ca2c662cb79636cf9b7b8e
    SMLicenseCertName: jfrog-artifactory
    Certificate: "-----BEGIN CERTIFICATE-----|  CERTIFICATE_MATCHING_DOMAIN_LINE_SEPARATORES_WITH|-----END     CERTIFICATE-----"
    CertificateKey: "-----BEGIN PRIVATE KEY-----|PRIVATE_KEY_MATCHING_DOMAIN_LINE_SEPARATORES_WITH`|`|-----END PRIVATE KEY-----"

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

## Maintenance

### Updating AMIs

The easiest way to create a mapping for a particular AMI is run the following:

    ```bash
    ami_name=<name_of_ami>
    regions=(us-east-2 us-east-1 us-west-2 us-west-1 eu-west-3 eu-west-2 eu-west-1 eu-central-1 eu-north-1 ap-northeast-2 ap-northeast-1 ap-southeast-2 ap-southeast-1 ca-central-1 ap-south-1 sa-east-1)
    owner=309956199498
    for i in "${regions[@]}"; do         echo "    $i:";         echo "       AMZNLINUXHVM: `aws --region $i ec2 describe-images --owners $owner --filters \"Name=name,Values=$ami_name\" 'Name=state,Values=available' --output json | jq -r '.Images | sort_by(.CreationDate) | last(.[]).ImageId'`";     done
    ```

This will output the proper list (Depending on your mapping of course):

    ```bash
    us-east-2:
       RHEL: ami-0cf433f9a817f63d3
    us-east-1:
       RHEL: ami-029c0fbe456d58bd1
    us-west-2:
       RHEL: ami-078a6a18fb73909b2
    us-west-1:
       RHEL: ami-07d8d14365439bc6e
    eu-west-3:
       RHEL: ami-018c55e9d34f949e9
    eu-west-2:
       RHEL: ami-0d8f9df7aa93d806e
    eu-west-1:
       RHEL: ami-065ec1e661d619058
    eu-central-1:
       RHEL: ami-06220be3176081cf0
    eu-north-1:
       RHEL: ami-a4fe74da
    ap-northeast-2:
       RHEL: ami-0708fd0ae9a663e02
    ap-northeast-1:
       RHEL: ami-0be4c0b05bbeb2afd
    ap-southeast-2:
       RHEL: ami-01448715c06d2edb5
    ap-southeast-1:
       RHEL: ami-02079c0159aade6b4
    ca-central-1:
       RHEL: ami-05508913d3360e9af
    ap-south-1:
       RHEL: ami-003b12a9a1ee83922
    sa-east-1:
       RHEL: ami-0102667d2046392a0
    ```
