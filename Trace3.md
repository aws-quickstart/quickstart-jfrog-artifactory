# quickstart-jfrog-artifactory-Trace3-Internal

This is a Quick Start to get Enterprise Production ready Artifactory deployed into your AWS environment. There are several parts to a Quickstart.

## Getting started

To complete a Quickstart you need the following items:

* Quickstart Word document (QuickStart Guide)
* Diagrams
* CloudFormation code
* Taskcat output of a successful deployment.

## Quickstart Word document

For this specific quickstart all documentation is being stored in [Box](https://trace3.box.com/s/e0infq2amuefkxqllbveq60xorpeee7f)
Note the Word document has the diagrams within it.

## Diagrams

Diagrams for this specific project are embedded within the project and also part of this [powerpoint](https://trace3.box.com/s/m97mbf4yfazdm8ruhmfsdd9x4lycukg6)
You will need to ensure you download the latest templates for the diagrams. Note: When using icons for services such as ECS, RDS, etc. they need to be within the AWS cloud box.

## CloudFormation code

Ensure Parameters fit the diction that is required. No Camel Case for them, need to end with a period and be spelled out where possible. This can take a considerable amount of effort during the approval page if you do not follow the proper way. DO not copy other CF templates as they are wrong, the templates within this should conform.

## Taskcat

Taskcat information can be found in the README.md.
We have extra .json examples within our internal repo. If you are debugging a stack, such as EKS you can deploy the whole stack, then just delete the "core" stack. Remeber, the name of the template will be somewhere in the name of the stack. Doing the core-workload will delete the helm deployment. You can then use the jfrog-artifactory-eks-core.json and fill in your variables. Use the Parameters of the core as a great starting point, as well as, the Outputs of the other stacks like core-infrastructure for the S3 key and properties. Then running taskcat with that file enabled will allow you to iterate much quicker than an entire stack creation/deletion.
