#!/usr/bin/env bash
echo $BASH_VERSION

# NOTE: macOS default shell is zsh so bash has been left behind as version 3
# Upgrade Bash to v5:
#   brew install bash
#
# This install new bash to /usr/local/bin/bash. You can check with 'which -a bash'
#
# macOS version of sed isn't POSIX so you should install GNU sed:
#   brew install gnu-sed
#
# Then follow the instruction and add the path to your .zshrc

if [[ -z "${S3_PREFIX}" ]]; then
  echo "Env var S3_PREFIX is not set! e.g. 'artifactory7/pro/v7125'"
  exit 1
fi

TEMPLATE_SOURCE_DIR="../templates"
TEMPLATE_OUTPUT_DIR="../templates/.transformed_output"

TEMPLATE_FILES=(
  jfrog-artifactory-ec2-existing-vpc.template.yaml
  jfrog-artifactory-ec2-master.template.yaml
  jfrog-artifactory-pro-ec2-existing-vpc-master.template.yaml
  jfrog-artifactory-pro-ec2-new-vpc-master.template.yaml
)

S3_BUCKET_NAME=${S3_BUCKET_NAME:="jfrog-aws-test"}
S3_BUCKET_REGION=${S3_BUCKET_REGION:="us-east-1"}

mkdir -p $TEMPLATE_OUTPUT_DIR

for file in "${TEMPLATE_FILES[@]}"; do
    echo "Copying $file to '$TEMPLATE_OUTPUT_DIR'"
    cp $TEMPLATE_SOURCE_DIR/$file $TEMPLATE_OUTPUT_DIR/$file

    echo "Removing !Sub values from TemplateURL"
    # Remove the 2 lines for !Sub in TemplateURL:
    # - S3Bucket: !If [UsingDefaultBucket, !Sub '${QsS3BucketName}-${AWS::Region}', !Ref 'QsS3BucketName']
    #   S3Region: !If [UsingDefaultBucket, !Ref 'AWS::Region', !Ref 'QsS3BucketRegion']
    sed -i 's/ *- S3Bucket: !If \[.*//g' $TEMPLATE_OUTPUT_DIR/$file
    sed -i 's/ *S3Region: !If \[.*//g' $TEMPLATE_OUTPUT_DIR/$file

    echo "Replacing variables from TemplateURL"
    # Replace the !Sub source
    # From:
    #   https://${S3Bucket}.s3.${S3Region}.${AWS::URLSuffix}/${QsS3KeyPrefix}templates/jfrog-artifactory-ec2-master.template.yaml
    # To:
    #   https://$QsS3BucketName.s3.$QsS3BucketRegion.${AWS::URLSuffix}/$QsS3KeyPrefix/templates/jfrog-artifactory-ec2-master.template.yaml
    sed -i "s/https:\/\/\${S3Bucket}\.s3\.\${S3Region}\.\${AWS::URLSuffix}\/\${QsS3KeyPrefix}\(.*template.*\)/https:\/\/${S3_BUCKET_NAME}\.s3\.${S3_BUCKET_REGION}\.\${AWS::URLSuffix\}\/${S3_PREFIX//\//\\/}\/\1/g" $TEMPLATE_OUTPUT_DIR/$file

    # Remove line with only spaces
    sed -i '/^[[:space:]]*$/d' $TEMPLATE_OUTPUT_DIR/$file

    echo "Removing newline from TemplateURL"
    # Remove the linebreak after !Sub for TemplateURL
    # From:
    #   TemplateURL: !Sub
    #     - https://${S3Bucket}.s3.${S3Region}.${AWS::URLSuffix}/${QsS3KeyPrefix}templates/jfrog-artifactory-ec2-master.template.yaml
    # To:
    #   TemplateURL: !Sub https://${S3Bucket}.s3.${S3Region}.${AWS::URLSuffix}/${QsS3KeyPrefix}templates/jfrog-artifactory-ec2-master.template.yaml
    sed -i -z 's/TemplateURL: !Sub\n *-/TemplateURL: !Sub/g' $TEMPLATE_OUTPUT_DIR/$file

    echo "Replacing 'QsS3KeyPrefix' value"
    # Replace Parameter reference with actual path
    # From:
    #   QsS3KeyPrefix: !Ref QsS3KeyPrefix
    # To:
    #   QsS3KeyPrefix: "artifactory7/$FOLDER_NAME/"
    sed -i "s/QsS3KeyPrefix: !Ref.*/QsS3KeyPrefix: \"${S3_PREFIX//\//\\/}\/\"/g" $TEMPLATE_OUTPUT_DIR/$file

    echo "Replacing 'QsS3BucketName' value"
    # Replace Parameter reference with actual value
    # From:
    #   !Ref QsS3BucketName or ${QsS3BucketName}
    # To:
    #   jfrog-aws-test
    sed -i -E "s/!Ref '?QsS3BucketName'?/${S3_BUCKET_NAME}/g" $TEMPLATE_OUTPUT_DIR/$file
    sed -i "s/\${QsS3BucketName}/${S3_BUCKET_NAME}/g" $TEMPLATE_OUTPUT_DIR/$file

    echo "Replacing 'QsS3BucketRegion' referenced value"
    # Replace Parameter reference with actual value
    # From:
    #   !Ref QsS3BucketRegion
    # To:
    #   us-east-1
    sed -i -E "s/!Ref '?QsS3BucketRegion'?/${S3_BUCKET_REGION}/g" $TEMPLATE_OUTPUT_DIR/$file

    echo "Replacing 'QSS3KeyPrefix' value"
    # Replace Parameter reference with actual path
    # From:
    #   QSS3KeyPrefix: !Sub '${QsS3KeyPrefix}submodules/quickstart-linux-bastion/'
    # To:
    #   QSS3KeyPrefix: !Sub 'artifactory7/$FOLDER_NAME/submodules/quickstart-linux-bastion/'
    sed -i "s/QSS3KeyPrefix: !Sub '\${QsS3KeyPrefix}\(.*\)'/QSS3KeyPrefix: '${S3_PREFIX//\//\\/}\/\1'/g" $TEMPLATE_OUTPUT_DIR/$file
done

cp $TEMPLATE_SOURCE_DIR/jfrog-artifactory-core-infrastructure.template.yaml $TEMPLATE_OUTPUT_DIR/jfrog-artifactory-core-infrastructure.template.yaml
cp $TEMPLATE_SOURCE_DIR/jfrog-artifactory-ec2-instance.template.yaml $TEMPLATE_OUTPUT_DIR/jfrog-artifactory-ec2-instance.template.yaml
