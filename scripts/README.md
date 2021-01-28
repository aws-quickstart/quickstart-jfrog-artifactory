## Prepare CFTs for AWS review

### Transform CFTs and upload to S3

* Set environment variables (`S3_BUCKET_NAME`, `S3_BUCKET_REGION`, `S3_PREFIX`)
* Run `transformed_templates_for_qs.sh` to replace parameters with hardcoded values.

```sh
$ export S3_BUCKET_NAME=jfrog-aws-test
$ export S3_BUCKET_REGION=us-east-1
$ export S3_PREFIX="artifactory7/pro/v7125"
$ ./transformed_templates_for_qs.sh
```

The transformed templates will be written to `templates/.transformed_output` directory.

Run `upload_to_s3.sh <S3 bucket name> <S3 prefix>` to upload the transformed templates to S3 bucket

```sh
$ ./upload_to_s3.sh jfrog-aws-test "artifactory7/pro/v7125"
```

### Test in CloudFormation

Create a new stack using the URL from the S3 bucket, e.g. `https://jfrog-aws-test.s3.amazonaws.com/artifactory7/pro/v7125/templates/jfrog-artifactory-pro-ec2-new-vpc-master.template.yaml`

Verify Artifactory and Xray are up and running.
