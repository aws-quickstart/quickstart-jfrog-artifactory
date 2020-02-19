artifactory-ami
=========

The main usage of this role is to configure a VM for artifactory install, with just a binary and the certificates in the keystore. Then clean the instance up, and shut down to be ready for consumption for the AWS Marketplace and AWS Quickstarts.

Requirements
------------

Ansible > 2.8

Role Variables
--------------

artifactory_version: The version of artifactory you wish to bake into this image.

Dependencies
------------

Ansible > 2.8

Example Playbook
----------------

Including an example of how to use your role (for instance, with variables passed in as parameters) is always nice for users too:

    - import_playbook: artifactory-ami.yml
      vars:
        artifactory_version: 7.0.0

Then artifactory-ami.yml would look like so:

    - hosts: localhost
      gather_facts: true
      become: true
      roles:
      - name: artifactory-ami
