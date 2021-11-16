## JFrog Artifactory and Xray on the AWS Cloud—Quick Start

For architectural details, step-by-step instructions, and customization options, see the [deployment guide](https://aws-quickstart.github.io/quickstart-jfrog-artifactory/).

To post feedback, submit feature ideas, or report bugs, use the **Issues** section of this GitHub repo.

To submit code for this Quick Start, see the [AWS Quick Start Contributor's Kit](https://aws-quickstart.github.io/).

## Upgrading RT and Xray Version

- Create a PR to merge latest AWS code to jfrog frok.
  [Create Pull Request](https://github.com/jfrog/quickstart-jfrog-artifactory/compare/main...aws-quickstart:main)
  Incase of merge conflicts commands used
  ```
  git stash
  git pull
  git pull https://github.com/aws-quickstart/quickstart-jfrog-artifactory.git main
  git checkout main
  git merge --no-ff aws-quickstart-main
  git push origin main
  ```
- Clone the repo and checkout new branch
  ```
  export VERSION=7.27.10 #Artifactory version
  git clone https://github.com/jfrog/quickstart-jfrog-artifactory.git
  cd quickstart-jfrog-artifactory
  git checkout main
  git pull
  git checkout -b $VERSION
  ```
- Update artifactoryVersion and xrayVersion in following files
  - templates/jfrog-artifactory-ec2-existing-vpc.template.yaml
  - templates/jfrog-artifactory-ec2-master.template.yaml
  - templates/jfrog-artifactory-pro-ec2-existing-vpc-master.template.yaml
  - templates/jfrog-artifactory-pro-ec2-new-vpc-master.template.yaml
- Run it locally to make sure code runs before extensive testing.
  This will not delete the stack make sure to delete it manually once verification is done.
  Empty the s3 bucket "tcat-qs-ec2"-<region> in aws before running the test to clean up old files.
  Create a .taskcat.yml file in home directory $HOME/.taskcat.yaml and set following parameters

  - KeyPairName
  - AccessCidr
  - RemoteAccessCidr
    Sample global taskcat file

  ```
  project:
  parameters:
    KeyPairName: <Key Pair Name>
    AccessCidr: <CIDR>
    RemoteAccessCidr: <CIDR>
  ```

  Use these commands

  ```
  make submodules
  cp .jfrog.taskcat.yml .taskcat.yml
  taskcat test run -n -l -t ent-existing-vpc-e1
  ```

  Once the stacks are up do basic validation.

- Fix taskcat, make any fixes as needed in this file. If parameters are removed or added, add it in the file
  ```
  mv .taskcat.yml .jfrog.taskcat.yml
  git checkout .taskcat.yml
  diff .taskcat.yml .jfrog.taskcat.yml
  ```
- Push changes into origin
  ```
  git status
  git add templates/\*
  git add .jfrog.taskcat.yml
  git status
  git commit -m "updated version to $VERSION"
  git push origin $VERSION
  ```
- For testing other scenarios go to [Jfrog pipelines](https://partnership.jfrog.io/ui/pipelines/myPipelines/default/AWS_Deployment_Pipeline?branch=aws-pipeline) and trigger the pipelines with following parameters.
  | Test | Region | TaskName | KeyPair | Rt_ver eg: 7.27.10 | Xray_ver eg:3.35.0 | Profile | BrachName eg: 7.27.10 |
  |---------------------|---------------|--------------------------------|----------------|-----------------------|--------------------|---------|-----------------------------|
  | RT Ent new VPC | us-east-1 | artifactory-enterprise-new-vpc | \<keypair name\> | \<Artifactory version\> | \<Xray version\> | default | \<newly created branch name\> |
  | RT pro new VPC | us-east-2 | artifactory-pro-new-vpc | \<keypair name\> | \<Artifactory version\> | \<Xray version\> | default | \<newly created branch name\> |
  | RT pro existing VPC | us-east-1 | artifactory-pro-existing-vpc | \<keypair name\> | \<Artifactory version\> | \<Xray version\> | default | \<newly created branch name\> |
  | RT Ent new VPC | us-gov-east-1 | artifactory-enterprise-new-vpc | \<keypair name\> | \<Artifactory version\> | \<Xray version\> | gov | \<newly created branch name\> |

- Wait for all stacks to come up successfully then validate using “basic validation” instructions given below.

- Create PR to main branch of jfrog fork for review

## Basic Validation

- Log into Artifactory change default password to \<password\>.
  Url for artifactory can be found in stack output.
- Create a new docker repo “test1”
- export DNS=<aws elb dns>
- docker login -u admin -p "\<password\>" $DNS
- docker pull alpine
- docker tag alpine $DNS/test1/alpine
- docker push $DNS/test1/alpine
- docker rmi alpine $DNS/test1/alpine
- docker pull $DNS/test1/alpine
- docker tag $DNS/test1/alpine alpine
- docker rmi $DNS/test1/alpine

> Note: If the cert is expired.Then you have to add it add the DNS to insecure-registry list in docker deamon config.
