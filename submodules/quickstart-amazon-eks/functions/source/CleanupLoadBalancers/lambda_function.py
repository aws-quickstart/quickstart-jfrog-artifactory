#  Copyright 2016 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
#  This file is licensed to you under the AWS Customer Agreement (the "License").
#  You may not use this file except in compliance with the License.
#  A copy of the License is located at http://aws.amazon.com/agreement/ .
#  This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, express or implied.
#  See the License for the specific language governing permissions and limitations under the License.

import boto3
import logging
from crhelper import CfnResource

logger = logging.getLogger(__name__)
helper = CfnResource(json_logging=True, log_level='DEBUG')


def delete_dependencies(sg_id, c):
    filters = [{'Name': 'ip-permission.group-id', 'Values': [sg_id]}]
    for sg in c.describe_security_groups(Filters=filters)['SecurityGroups']:
        for p in sg['IpPermissions']:
            if 'UserIdGroupPairs' in p.keys():
                if sg_id in [x['GroupId'] for x in p['UserIdGroupPairs']]:
                    try:
                        c.revoke_security_group_ingress(GroupId=sg['GroupId'], IpPermissions=[p])
                    except Exception as e:
                        logger.error("ERROR: %s %s" % (sg['GroupId'], str(e)))
    filters = [{'Name': 'egress.ip-permission.group-id', 'Values': [sg_id]}]
    for sg in c.describe_security_groups(Filters=filters)['SecurityGroups']:
        for p in sg['IpPermissionsEgress']:
            if 'UserIdGroupPairs' in p.keys():
                if sg_id in [x['GroupId'] for x in p['UserIdGroupPairs']]:
                    try:
                        c.revoke_security_group_egress(GroupId=sg['GroupId'], IpPermissions=[p])
                    except Exception as e:
                        logger.error("ERROR: %s %s" % (sg['GroupId'], str(e)))
    filters = [{'Name': 'group-id', 'Values': [sg_id]}]
    for eni in c.describe_network_interfaces(Filters=filters)['NetworkInterfaces']:
        try:
            c.delete_network_interface(NetworkInterfaceId=eni['NetworkInterfaceId'])
        except Exception as e:
            logger.error("ERROR: %s %s" % (eni['NetworkInterfaceId'], str(e)))


@helper.delete
def delete_handler(event, _):
    tag_key = "kubernetes.io/cluster/%s" % event["ResourceProperties"]["ClusterName"]
    lb_types = [
        ["elb", "LoadBalancerName", "LoadBalancerNames", "LoadBalancerDescriptions", "LoadBalancerName"],
        ["elbv2", "LoadBalancerArn", "ResourceArns", "LoadBalancers", "ResourceArn"]
    ]
    for lt in lb_types:
        elb = boto3.client(lt[0])
        lbs = []
        response = elb.describe_load_balancers()
        while True:
            lbs += [l[lt[1]] for l in response[lt[3]]]
            if "NextMarker" in response.keys():
                response = elb.describe_load_balancers(Marker=response["NextMarker"])
            else:
                break
        lbs_to_remove = []
        if lbs:
            #Split LB list into groups of 'size' items.
            size = 20
            lb_groups = (lbs[pos:pos + size] for pos in range(0, len(lbs), size))
            for lb_group in lb_groups:
                lb_group = elb.describe_tags(**{lt[2]: lb_group})["TagDescriptions"]
                for tags in lb_group:
                    for tag in tags['Tags']:
                        if tag["Key"] == tag_key and tag['Value'] == "owned":
                            lbs_to_remove.append(tags[lt[4]])
        if lbs_to_remove:
            for lb in lbs_to_remove:
                print("removing elb %s" % lb)
                elb.delete_load_balancer(**{lt[1]: lb})
    ec2 = boto3.client('ec2')
    response = ec2.describe_tags(Filters=[
        {'Name': 'tag:%s' % tag_key, 'Values': ['owned']},
        {'Name': 'resource-type', 'Values': ['security-group']}
    ])
    for t in [r['ResourceId'] for r in response['Tags']]:
        try:
            ec2.delete_security_group(GroupId=t)
        except ec2.exceptions.ClientError as e:
            if 'DependencyViolation' in str(e):
                print("Dependency error on %s" % t)
                delete_dependencies(t, ec2)
            else:
                raise


def lambda_handler(event, context):
    helper(event, context)
