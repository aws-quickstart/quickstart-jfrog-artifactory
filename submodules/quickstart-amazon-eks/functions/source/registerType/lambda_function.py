import boto3
import logging
from crhelper import CfnResource
from time import sleep
import json

logger = logging.getLogger(__name__)
helper = CfnResource(json_logging=True, log_level='DEBUG')
cfn = boto3.client('cloudformation')


def stabilize(token):
    p = cfn.describe_type_registration(RegistrationToken=token)
    while p['ProgressStatus'] == "IN_PROGRESS":
        sleep(5)
        p = cfn.describe_type_registration(RegistrationToken=token)
    if p['ProgressStatus'] == 'FAILED':
        if 'to finish before submitting another deployment request for ' not in p['Description']:
            raise Exception(p['Description'])
        return None
    return p['TypeVersionArn']


@helper.create
@helper.update
def register(event, _):
    logger.error(f"event: {json.dumps(event)}")
    kwargs = {
       "Type": 'RESOURCE',
       "TypeName": event['ResourceProperties']['TypeName'],
       "SchemaHandlerPackage": event['ResourceProperties']['SchemaHandlerPackage'],
        "LoggingConfig": {
            "LogRoleArn": event['ResourceProperties']['LogRoleArn'],
            "LogGroupName": event['ResourceProperties']['LogGroupName']
        },
        "ExecutionRoleArn": event['ResourceProperties']['ExecutionRoleArn']
    }
    response = cfn.register_type(**kwargs)
    version_arn = stabilize(response['RegistrationToken'])
    if version_arn:
        cfn.set_type_default_version(Arn=version_arn)
    return version_arn


@helper.delete
def deregister(event, _):
    type_name = event['ResourceProperties']['TypeName']
    versions = cfn.list_type_versions(Type='RESOURCE', TypeName=type_name)['TypeVersionSummaries']
    if len(versions) > 1:
        cfn.deregister_type(Arn=event['PhysicalResourceId'])
    else:
        cfn.deregister_type(Type='RESOURCE', TypeName=type_name)


def lambda_handler(event, context):
    helper(event, context)
