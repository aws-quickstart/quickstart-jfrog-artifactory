## Prepare CFTs for AWS review

### Transform CFTs and upload to S3

#### Transform CFTs and replace QuickStart parameter values with hardcoded values

* Set environment variables (`S3_BUCKET_NAME`, `S3_BUCKET_REGION`, `S3_PREFIX`)
* Run `transformed_templates_for_qs.sh` to replace parameters with hardcoded values.

```sh
$ export S3_BUCKET_NAME=jfrog-aws-test
$ export S3_BUCKET_REGION=us-east-1
$ export S3_PREFIX="artifactory7/pro/v7125"
$ ./transformed_templates_for_qs.sh
```

The transformed templates will be written to `templates/.transformed_output` directory.

#### Remove QuickStart parameters from root templates

Manually remove the following parameters (easier than regex!) from the transformed root templates (`jfrog-artifactory-pro-ec2-new-vpc-master.template.yaml` and `jfrog-artifactory-pro-ec2-existing-vpc-master.template.yaml`) so they don't appear in CloudFormation console:
- `QsS3BucketName`
- `QsS3KeyPrefix`
- `QsS3BucketRegion`

They appears in `Metadata` and `Parameters` sections

#### Upload templates to S3

Run `upload_to_s3.sh <S3 bucket name> <S3 prefix>` to upload the transformed templates to S3 bucket

```sh
$ ./upload_to_s3.sh
```

Override S3 bucket name and prefix using arguments:

```sh
$ ./upload_to_s3.sh jfrog-aws-test "artifactory7/pro/v7125"
```

### Test in CloudFormation

Create a new stack using the URL from the S3 bucket, e.g. `https://jfrog-aws-test.s3.amazonaws.com/artifactory7/pro/v7125/templates/jfrog-artifactory-pro-ec2-new-vpc-master.template.yaml`

Enter the appropriate parameters e.g.
```
Stack name: artifactory7 # append with some unique number
SSH key name: <your key>
Permitted IP range: 0.0.0.0/0
Remote access CIDR:	0.0.0.0/0
SmLicenseCertName: jfrog-artifactory
Artifactory server name:	artifactory
masterkey: FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF
Various passwords: artifactory
```

Verify the stack is created successfully, and both Artifactory and Xray are up and running.

#### Validate RT

For basic validation follow these steps:
- Log into Artifactory or JCR, follow welcome steps, change password to “Password1”
- Create a new docker repo “test1”
- Now on command line follow these commands for basic validation:

```sh
export DNS=<Internet facing ELB DNS>
docker login -u admin -p Password1 $DNS
docker pull alpine
docker tag alpine $DNS/test1/alpine
docker push $DNS/test1/alpine # if you get error “ace0eda3e3be: Retrying in 4 seconds”, make sure repo was created.
docker rmi alpine $DNS/test1/alpine
docker rmi -f <image id>
docker pull $DNS/test1/alpine
docker tag $DNS/test1/alpine alpine
docker rmi $DNS/test1/alpine
```

### How to debug

Logs (not all) are sent to CloudWatch so that's the best starting point. They are under the log groups:
- `/artifactory/instances/<instance ID>`
- `/xray/instances/<instance ID>`

To SSH into the EC2 instance when there is no bastion, run this taskcat test (`create-bastion-with-existing-vpc`) to create the bastion stack. Adjust network parameters to suit the region. Change the `RemoteAccessCIDR` parameter to match your ISP public IP. *Don't* use `0.0.0.0/0`!

To decrypt Ansible playbook:
```sh
$ sudo su
$ source /root/venv/bin/activate
$ ansible-vault decrypt /root/.jfrog_ami/artifactory.yml --vault-id /root/.vault_pass.txt
```

### Update JFrog-Cloud-Installers repo

Once templates are verified, copy the templates to the [JFrog-Cloud-Installers repo](https://github.com/jfrog/JFrog-Cloud-Installers).

Open a PR and add reviewers.
