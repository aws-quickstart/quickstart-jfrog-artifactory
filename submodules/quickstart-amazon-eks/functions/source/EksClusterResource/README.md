# AWSQS::EKS::Cluster
***beta***
 
An AWS CloudFormation resource provider for modelling Amazon EKS clusters. 
It provides some additional functionality to the built-in resource provider:

* Manage `aws-auth` ConfigMap from within CloudFormation.
* Support for `EndpointPublicAccess`, `EndpointPrivateAccess` and 
`PublicAccessCidrs` features.
* Support for enabling control plane logging to CloudWatch logs.   

Properties and available attributes (ReadOnlyProperties) are documented in 
the [schema](./awsqs-eks-cluster.json).

## Installation
```bash
aws cloudformation create-stack \
  --stack-name awsqs-eks-cluster-resource \
  --capabilities CAPABILITY_NAMED_IAM \
  --template-url https://s3.amazonaws.com/aws-quickstart/quickstart-amazon-eks-cluster-resource-provider/deploy.template.yaml \
  --region us-west-2 \
  --parameters ParameterKey=CreateClusterAccessRole,ParameterValue='true' # set to false if you have already deployed once in another region
```
A [template](./deploy.template.yaml) is provided to make deploying the resource into 
an account easy. Set `CreateClusterAccessRole` to `false` if the execution role has 
already been created (if you've previously added the resource to another region in the same account).

Example usage:

```yaml
AWSTemplateFormatVersion: "2010-09-09"
Parameters:
  SubnetIds:
    Type: "List<AWS::EC2::Subnet::Id>"
  SecurityGroupIds:
    Type: "List<AWS::EC2::SecurityGroup::Id>"
Resources:
  # EKS Cluster
  myCluster:
    Type: "AWSQS::EKS::Cluster"
    Properties:
      RoleArn: !GetAtt serviceRole.Arn
      ResourcesVpcConfig:
        SubnetIds: !Ref SubnetIds
        SecurityGroupIds: !Ref SecurityGroupIds
        EndpointPrivateAccess: true
        EndpointPublicAccess: true
      EnabledClusterLoggingTypes: ["audit"]
      KubernetesApiAccess:
        Users:
          - Arn: "arn:${AWS::Partition}:iam::${AWS::AccountId}:user/my-user"
            Username: "CliUser"
            Groups: ["system:masters"]
  serviceRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal: { Service: eks.amazonaws.com }
            Action: sts:AssumeRole
      Path: "/"
      ManagedPolicyArns:
        - !Sub 'arn:${AWS::Partition}:iam::aws:policy/AmazonEKSClusterPolicy'
        - !Sub 'arn:${AWS::Partition}:iam::aws:policy/AmazonEKSServicePolicy'
```
