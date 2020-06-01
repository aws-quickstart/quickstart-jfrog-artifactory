import boto3
import logging
from crhelper import CfnResource

logger = logging.getLogger(__name__)
helper = CfnResource(json_logging=True, log_level='DEBUG')


@helper.delete
def delete_objects(event, _c):
    bucket_name = event["ResourceProperties"]["Bucket"]
    s3 = boto3.client('s3')

    logger.info('Getting objects...')
    objects = []
    kwargs = {"Bucket": bucket_name}
    while True:
        versions = s3.list_object_versions(**kwargs)
        if 'Versions' in versions.keys():
            for v in versions['Versions']:
                objects.append({'Key': v['Key'], 'VersionId': v['VersionId']})
        if 'DeleteMarkers' in versions.keys():
            for v in versions['DeleteMarkers']:
                objects.append({'Key': v['Key'], 'VersionId': v['VersionId']})
        if versions['IsTruncated']:
            if versions.get('NextKeyMarker', 'null') != 'null':
                kwargs["KeyMarker"] = versions['NextKeyMarker']
            else:
                if kwargs.get("KeyMarker"):
                    del kwargs["KeyMarker"]
            if versions.get('NextVersionIdMarker', 'null') != 'null':
                kwargs["VersionIdMarker"] = versions['NextVersionIdMarker']
            else:
                if kwargs.get("VersionIdMarker"):
                    del kwargs["VersionIdMarker"]
        else:
            break
    if objects:
        # delete objects in batches of 1000
        for i in range(0, len(objects), 1000):
            s3.delete_objects(
                Bucket=bucket_name,
                Delete={'Objects': objects[i:i + 1000]}
            )


def lambda_handler(event, context):
    helper(event, context)
