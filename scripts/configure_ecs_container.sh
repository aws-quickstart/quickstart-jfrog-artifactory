#/bin/bash
whoami > /tmp/startup.log
echo ${YQ_VERSION} >> /tmp/startup.log
echo "downloading YQ" >> /tmp/startup.log
wget "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -O /tmp/yq
echo "Making it Executable"
chmod +x /tmp/yq
echo "Creating System YAML file" >> /tmp/startup.log
/tmp/yq n shared.node.primary ${HA_IS_PRIMARY} | tee /tmp/system.yaml
/tmp/yq w -i /tmp/system.yaml configVersion 1
/tmp/yq w -i /tmp/system.yaml -- shared.extraJavaOpts "${EXTRA_JAVA_OPTIONS}"
/tmp/yq w -i /tmp/system.yaml shared.node.haEnabled true
/tmp/yq w -i /tmp/system.yaml shared.database.type ${DB_TYPE}
/tmp/yq w -i /tmp/system.yaml shared.database.driver ${DB_DRIVER}
/tmp/yq w -i /tmp/system.yaml shared.database.url ${DB_URL}
/tmp/yq w -i /tmp/system.yaml shared.database.username ${DB_USER}
/tmp/yq w -i /tmp/system.yaml shared.database.password ${DB_PASSWORD}

echo "Moving file to proper location" >> /tmp/startup.log
mv /tmp/system.yaml /var/opt/jfrog/artifactory/etc/

cat /var/opt/jfrog/artifactory/etc/system.yaml >> /tmp/startup.log
echo "Creating Master Key" >> /tmp/startup.log
echo ${ARTIFACTORY_MASTER_KEY} > /var/opt/jfrog/artifactory/etc/security/master.key