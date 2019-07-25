create:
	aws cloudformation create-stack --stack-name test --template-body file://$(pwd)/templates/jfrog-artifactory-ec2-new-vpc.template --parameters $(cat .ignore/params) --capabilities CAPABILITY_IAM

delete:
	aws cloudformation delete-stack --stack-name test

test: lint
	taskcat -c ci/config.yml

lint:
	taskcat -l -c ci/config.yml
