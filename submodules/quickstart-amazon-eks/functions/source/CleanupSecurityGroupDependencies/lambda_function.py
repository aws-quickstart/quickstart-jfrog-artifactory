#  Copyright 2016 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
#  This file is licensed to you under the AWS Customer Agreement (the "License").
#  You may not use this file except in compliance with the License.
#  A copy of the License is located at http://aws.amazon.com/agreement/ .
#  This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, express or implied.
#  See the License for the specific language governing permissions and limitations under the License.

import logging
import boto3
from time import sleep
from crhelper import CfnResource

logger = logging.getLogger(__name__)
helper = CfnResource(json_logging=True, log_level='DEBUG')


def get_attachment_id_for_eni(eni):
    try:
        return eni['Attachment']['AttachmentId']
    except KeyError:
        return None


def delete_dependencies(sg_id, c):
    complete = True
    filters = [{'Name': 'ip-permission.group-id', 'Values': [sg_id]}]
    for sg in c.describe_security_groups(Filters=filters)['SecurityGroups']:
        for p in sg['IpPermissions']:
            if 'UserIdGroupPairs' in p.keys():
                if sg_id in [x['GroupId'] for x in p['UserIdGroupPairs']]:
                    try:
                        c.revoke_security_group_ingress(GroupId=sg['GroupId'], IpPermissions=[p])
                    except Exception as e:
                        complete = False
                        logger.error("ERROR: %s %s" % (sg['GroupId'], str(e)))
    filters = [{'Name': 'egress.ip-permission.group-id', 'Values': [sg_id]}]
    for sg in c.describe_security_groups(Filters=filters)['SecurityGroups']:
        for p in sg['IpPermissionsEgress']:
            if 'UserIdGroupPairs' in p.keys():
                if sg_id in [x['GroupId'] for x in p['UserIdGroupPairs']]:
                    try:
                        c.revoke_security_group_egress(GroupId=sg['GroupId'], IpPermissions=[p])
                    except Exception as e:
                        complete = False
                        logger.error("ERROR: %s %s" % (sg['GroupId'], str(e)))
    filters = [{'Name': 'group-id', 'Values': [sg_id]}]
    for eni in c.describe_network_interfaces(Filters=filters)['NetworkInterfaces']:
        try:
            attachment_id = get_attachment_id_for_eni(eni)
            if attachment_id:
                c.detach_network_interface(AttachmentId=attachment_id, Force=True)
                sleep(5)
            c.delete_network_interface(NetworkInterfaceId=eni['NetworkInterfaceId'])
        except Exception as e:
            complete = False
            logger.error("ERROR: %s %s" % (eni['NetworkInterfaceId'], str(e)))
    return complete


@helper.delete
def delete_handler(event, _):
    ec2 = boto3.client('ec2')
    for sg_id in event["ResourceProperties"]["SecurityGroups"]:
        retries = 5
        while True:
            if delete_dependencies(sg_id, ec2):
                break
            if retries == 0:
                logger.error(f"failed to delete {sg_id} dependencies after 5 retries")
                break
            retries -= 1
            sleep(15)


def lambda_handler(event, context):
    helper(event, context)
