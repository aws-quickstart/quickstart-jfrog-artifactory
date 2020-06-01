import json
import boto3
import subprocess
import shlex
import random
import re
from crhelper import CfnResource
import logging
import string
from datetime import datetime
import requests


logger = logging.getLogger(__name__)
helper = CfnResource(json_logging=True, log_level='DEBUG')

try:
    s3_client = boto3.client('s3')
    kms_client = boto3.client('kms')
    valid_url_schemes = re.compile(r'^(?:http|https|s3)://')
    s3_scheme = re.compile(r'^s3://.+/.+')
except Exception as e:
    helper.init_failure(e)


def rand_string(l):
    return ''.join(random.choice(string.ascii_lowercase) for _ in range(l))


def run_command(command):
    logger.debug("executing command: %s" % command)
    err = None
    output = None
    try:
        output = subprocess.check_output(shlex.split(command), stderr=subprocess.STDOUT).decode("utf-8")
        logger.debug(output)
    except subprocess.CalledProcessError as exc:
        logger.debug("Command failed with exit code %s, stderr: %s" % (exc.returncode, exc.output.decode("utf-8")))
        err = Exception(exc.output.decode("utf-8"))
    if err:
        try:
            if "Error: context deadline exceeded" in err:
                logger.error(f'retying command "{command}" as it failed with error: {err}')
                return run_command(command)
            else:
                raise err
        except TypeError:
            # Not iterable. (Simply raise the error)
            raise err
    else:
        return output


def create_kubeconfig(cluster_name):
    run_command(f"aws eks update-kubeconfig --name {cluster_name} --alias {cluster_name}")
    run_command(f"kubectl config use-context {cluster_name}")


def parse_install_output(output):
    data = {}
    resource_type = ""
    resources_block = False
    count = 0
    for line in output.split('\n'):
        if line.startswith("NAME:"):
            data['Name'] = line.split()[1]
        elif line.startswith("NAMESPACE:"):
            data['Namespace'] = line.split()[1]
        elif line == 'RESOURCES:':
            resources_block = True
        elif line == 'NOTES:':
            resources_block = False
        if resources_block:
            if line.startswith('==>'):
                count = 0
                if '==> MISSING' in line:
                    resource_type = ""
                else:
                    resource_type = line.split()[1].split('/')[1].replace('(related)', '')
            elif resource_type and not line.startswith('NAME') and line and ', Resource=' not in line:
                data[resource_type + str(count)] = line.split()[0]
                count += 1
            elif ', Resource=' in line and not resource_type:
                resource_type = line.split()[1].replace('Resource=', '')
                count = get_next_index(data, resource_type)
                data[resource_type + str(count)] = line.split()[2]
    return data


def get_next_index(data, resource_type):
    index = 0
    for t in data.keys():
        if t.startswith(resource_type) and t[len(resource_type):].isnumeric():
            index = int(t[len(resource_type):]) + 1
    return index


def write_values(manifest, path):
    f = open(path, "w")
    f.write(manifest)
    f.close()


def truncate(response_data):
    truncated = False
    while len(json.dumps(response_data)) > 3000:
        truncated = True
        response_data.pop(list(response_data.keys())[-1])
    response_data["Truncated"] = truncated
    return response_data


def helm_init(event):
    try:
        helper.Data.update({"StartTimestamp": event['CrHelperData']['StartTimestamp']})
    except KeyError:
        helper.Data.update({"StartTimestamp": str(datetime.now().timestamp())})
    physical_resource_id = None
    create_kubeconfig(event['ResourceProperties']['ClusterName'])
    run_command("helm --home /tmp/.helm init --client-only")
    repo_name = ''
    if 'Chart' in event['ResourceProperties'].keys():
        repo_name = event['ResourceProperties']['Chart'].split('/')[0]

    if "Name" in event['ResourceProperties'].keys():
        physical_resource_id = event["ResourceProperties"]["Name"]
    elif "PhysicalResourceId" in event.keys():
        physical_resource_id = event["PhysicalResourceId"]

    if "RepoUrl" in event['ResourceProperties'].keys():
        run_command("helm repo add %s %s --home /tmp/.helm" % (repo_name, event['ResourceProperties']["RepoUrl"]))
    if "Namespace" in event['ResourceProperties'].keys():
        namespace = event['ResourceProperties']["Namespace"]
        k8s_context = run_command("kubectl config current-context")
        run_command("kubectl config set-context %s --namespace=%s" % (k8s_context, namespace))
    run_command("helm --home /tmp/.helm repo update")
    return physical_resource_id


def http_get(url):
    try:
        response = requests.get(url)
    except requests.exceptions.RequestException as e:
        raise RuntimeError(f"Failed to fetch CustomValueYaml url {url}: {e}")
    if response.status_code != 200:
        raise RuntimeError(
            f"Failed to fetch CustomValueYaml url {url}: [{response.status_code}] "
            f"{response.reason}"
        )
    return response.text


def s3_get(url):
    try:
        return str(s3_client.get_object(
            Bucket=url.split('/')[2], Key="/".join(url.split('/')[3:])
        )['Body'].read())
    except Exception as e:
       raise RuntimeError(f"Failed to fetch CustomValueYaml {url} from S3. {e}")


def build_flags(properties, request_type="Create"):
    internal_values = ""
    if properties.get("ValueYaml"):
        write_values(properties["ValueYaml"], '/tmp/internalValues.yaml')
        internal_values = "-f /tmp/internalValues.yaml"
    custom_values = ""
    if properties.get("CustomValueYaml"):
        if not re.match(valid_url_schemes, properties["CustomValueYaml"]):
            raise ValueError()
        if re.match(s3_scheme, properties["CustomValueYaml"]):
            custom_value_yaml = s3_get(properties["CustomValueYaml"])
        else:
            custom_value_yaml = http_get(properties["CustomValueYaml"])
        write_values(custom_value_yaml, '/tmp/customValues.yaml')
        custom_values = "-f /tmp/customValues.yaml"
    set_vals = ""
    if properties.get("Values"):
        values = properties['Values']
        set_vals = " ".join(["--set %s=%s" % (k, values[k]) for k in values.keys()])
    version = ""
    if properties.get("Version"):
        version = "--version %s" % properties['Version']
    name = ""
    if properties.get("Name") and request_type != "Update":
        name = "--name %s" % properties['Name']
    if properties.get("ChartBucket") and properties.get("ChartKey"):
        properties['Chart'] = '/tmp/chart.tgz'
        chart = s3_client.get_object(Bucket=properties["ChartBucket"], Key=properties["ChartKey"])['Body'].read()
        f = open("/tmp/chart.tgz", "wb")
        f.write(chart)
        f.close()
    return "%s %s %s %s %s %s" % (properties['Chart'], internal_values, custom_values, set_vals, version, name)


def _trim_event_for_poll(event):
    needed_keys = ['Chart', 'RepoUrl', 'Namespace', 'ClusterName', 'TimeoutMinutes']
    trimmable = []
    for prop in event['ResourceProperties'].keys():
        if prop not in needed_keys:
            trimmable.append(prop)
    for key in trimmable:
        del event['ResourceProperties'][key]
    return event


@helper.create
def create(event, _):
    helm_init(event)
    cmd = "helm --home /tmp/.helm install %s" % build_flags(event['ResourceProperties'])
    output = run_command(cmd)
    response_data = parse_install_output(output)
    physical_resource_id = response_data["Name"]
    helper._event = _trim_event_for_poll(helper._event)
    return physical_resource_id


@helper.update
def update(event, _):
    physical_resource_id = helm_init(event)
    cmd = "helm --home /tmp/.helm upgrade %s %s" % (
        physical_resource_id, build_flags(event['ResourceProperties'], event["RequestType"]))
    output = run_command(cmd)
    response_data = parse_install_output(output)
    helper.Data.update(response_data)
    helper._event = _trim_event_for_poll(helper._event)
    return physical_resource_id


@helper.delete
def delete(event, _):
    physical_resource_id = helm_init(event)
    if not re.search(r'^[0-9]{4}/[0-9]{2}/[0-9]{2}/\[\$LATEST\][a-f0-9]{32}$', physical_resource_id):
        try:
            run_command("helm delete --home /tmp/.helm --purge %s" % physical_resource_id)
        except Exception as exc:
            if 'release: "%s" not found' % physical_resource_id in str(exc):
                logger.warning("release already gone, or never existed")
            elif 'invalid release name' in str(exc):
                logger.warning("release name invalid, either creation failed, or response not received by "
                               "CloudFormation")
            else:
                raise
    else:
        logger.warning("physical_resource_id is not a helm release, assuming there is nothing to delete")


@helper.poll_create
@helper.poll_update
def poll_create_update(event, _):
    helm_init(event)
    release_name = helper.Data["PhysicalResourceId"]
    cmd = "helm --home /tmp/.helm status %s" % release_name
    output = run_command(cmd)
    response_data = parse_install_output(output)
    ns = event['ResourceProperties']["Namespace"]
    unready = []
    for t in response_data.keys():
        k8s_type = t.rstrip(string.digits)
        if k8s_type.lower() in ["pod"]:
            k8s_name = response_data[t]
            output = run_command("kubectl get -o json -n %s %s/%s" % (ns, k8s_type, k8s_name))
            logger.debug(output)
            status = json.loads(output)["status"]
            if status['phase'] == 'Pending':
                msg = "%s/%s" % (k8s_type, k8s_name)
                unready.append(msg)
            if status['phase'] != "Succeeded":
                for s in status.get("containerStatuses", [{"ready": False}]):
                    if not s["ready"]:
                        msg = "%s/%s" % (k8s_type, k8s_name)
                        unready.append(msg)
    if unready:
        return poll_timeout(event, unready, release_name)
    helper.Data.update(truncate(response_data))
    return release_name


def poll_timeout(event, unready, release_name):
    start_time = datetime.fromtimestamp(float(helper.Data["StartTimestamp"]))
    total_duration_seconds = (datetime.now() - start_time).total_seconds()
    if 'TimeoutMinutes' not in event['ResourceProperties'].keys():
        timeout = 56 * 60
    else:
        timeout = int(event['ResourceProperties']['TimeoutMinutes']) * 60
    if total_duration_seconds >= timeout:
        logger.error("Polling about to timeout, sending failure to cloudformation")

        helper.PhysicalResourceId = release_name
        raise Exception("the following kubernetes resources were not ready before the timeout %s" % unready)
    return None


def lambda_handler(event, context):
    helper(event, context)
