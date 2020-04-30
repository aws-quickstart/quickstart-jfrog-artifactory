import boto3
import logging
from crhelper import CfnResource
from time import sleep

logger = logging.getLogger(__name__)
helper = CfnResource(json_logging=True, log_level='DEBUG')

ROLE_NAME = "CloudFormation-Kubernetes-VPC"
ASSUME_ROLE_POLICY_DOCUMENT = """{
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
}"""
POLICIES = [
    'arn:{}:iam::aws:policy/service-role/AWSLambdaENIManagementAccess',
    'arn:{}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole'
]


@helper.create
def create_role(event, _c):
    iam = boto3.client('iam')

    try:
        iam.create_role(
            RoleName=ROLE_NAME,
            AssumeRolePolicyDocument=ASSUME_ROLE_POLICY_DOCUMENT
        )
    except iam.exceptions.EntityAlreadyExistsException:
        print("Role already exists")
    retries = 0
    for p in POLICIES:
        while True:
            try:
                iam.attach_role_policy(RoleName=ROLE_NAME, PolicyArn=p.format(event['ResourceProperties']['Partition']))
                break
            except iam.exceptions.NoSuchEntityException:
                if retries > 20:
                    raise Exception("Failed to attach policy {} to role {}".format(
                        p.format(event['ResourceProperties']['Partition']),
                        ROLE_NAME
                    ))
                retries += 1
                sleep(5)
    return ROLE_NAME


def lambda_handler(event, context):
    helper(event, context)
