.PHONY: help run submodules
USEVENV ?= true
VENV=$${PWD}/venv
VENVBIN=${VENV}/bin
SHELL ?= /bin/bash

test: lint submodules
ifeq ($(USEVENV), true)
	$(MAKE) venv
	${VENVBIN}/taskcat test run -n -l
else
	taskcat test run -n -l
endif


venv/bin/python3:
		python3 -m venv ${VENV}

venv/bin/taskcat: venv/bin/python3
	${VENVBIN}/pip3 install taskcat

venv/bin/aws: venv/bin/python3
	${VENVBIN}/pip3 install awscli

venv: venv/bin/taskcat venv/bin/aws

submodules:
	git submodule init
	git submodule update --remote --recursive
	git submodule foreach --recursive 'git submodule init'
	git submodule foreach --recursive 'git submodule update --remote --recursive'

help:
	@echo   "make test  : executes ${VENVBIN}/taskcat"
	@echo   "if running in a container without venv please set USEVENV to false"


create: venv
	${VENVBIN}/aws cloudformation create-stack --stack-name test --template-body file://$(pwd)/templates/jfrog-artifactory-ec2-new-vpc.template --parameters $(cat .ignore/params) --capabilities CAPABILITY_IAM

delete: venv
	${VENVBIN}/aws cloudformation delete-stack --stack-name test

.ONESHELL:

lint:
ifeq ($(USEVENV), true)
	$(MAKE) venv
	time ${VENVBIN}/taskcat lint
else
	time taskcat lint
endif

public_repo: venv
	${VENVBIN}/taskcat -c theflash/ci/config.yml -u
	#https://${VENVBIN}/taskcat-tag-quickstart-jfrog-artifactory-c2fa9d34.s3-us-west-2.amazonaws.com/quickstart-jfrog-artifactory/templates/jfrog-artifactory-ec2-master.template
	#curl https://${VENVBIN}/taskcat-tag-quickstart-jfrog-artifactory-7008506c.s3-us-west-2.amazonaws.com/quickstart-jfrog-artifactory/templates/jfrog-artifactory-ec2-master.template

get_public_dns: venv
	${VENVBIN}/aws elb describe-load-balancers | jq '.LoadBalancerDescriptions[]| .CanonicalHostedZoneName'

get_bastion_ip: venv
	${VENVBIN}/aws ec2 describe-instances | jq '.[] | select(.[].Instances[].Tags[].Value == "LinuxBastion") '

clean:

realclean:
	rm -fr ${VENV} submodules
