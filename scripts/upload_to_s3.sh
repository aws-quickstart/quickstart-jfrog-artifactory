#!/usr/bin/env bash
echo $BASH_VERSION

SCRIPT_DIR=$(dirname "$0")

S3_BUCKET_NAME=${1:-"jfrog-aws-test"}
S3_PREFIX=${2:-"artifactory7/pro/v7125"}

TEMPLATE_OUTPUT_DIR=$SCRIPT_DIR/../templates/.transformed_output

if [ ! -d "$TEMPLATE_OUTPUT_DIR" ]; then
  echo "$TEMPLATE_OUTPUT_DIR directory doesn't exist. Run transformed_templates_for_qs.sh first"
  exit 1
fi

echo "Syncing directories with S3 bucket '$S3_BUCKET_NAME' with prefix '$S3_PREFIX'"
aws s3 sync "$SCRIPT_DIR/../cloudInstallerScripts" s3://$S3_BUCKET_NAME/$S3_PREFIX/cloudInstallerScripts
aws s3 sync "$SCRIPT_DIR/../submodules" s3://$S3_BUCKET_NAME/$S3_PREFIX/submodules
aws s3 sync "$TEMPLATE_OUTPUT_DIR" s3://$S3_BUCKET_NAME/$S3_PREFIX/templates
