#/bin/bash
startup_log=/tmp/startup.log
binarystore_temp=/tmp/binarystore.xml
system_temp=/tmp/system.yaml
YQ=/tmp/yq

whoami > ${startup_log}
echo ${YQ_VERSION} >> ${startup_log}
echo "downloading YQ" | tee ${startup_log}
wget "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64" -O ${YQ}
echo "Making it Executable"  | tee ${startup_log}
chmod +x ${YQ}
echo "Creating System YAML file" | tee ${startup_log}
${YQ} n shared.node.primary ${HA_IS_PRIMARY} | tee ${system_temp}
${YQ} w -i ${system_temp} configVersion 1
${YQ} w -i ${system_temp} -- shared.extraJavaOpts "${EXTRA_JAVA_OPTIONS}"
if [ -z "$HA_ENABLED" ]
then
  ${YQ} w -i ${system_temp} shared.node.haEnabled true
else
  ${YQ} w -i ${system_temp} shared.node.haEnabled $HA_ENABLED
fi

${YQ} w -i ${system_temp} shared.database.type ${DB_TYPE}
${YQ} w -i ${system_temp} shared.database.driver ${DB_DRIVER}
${YQ} w -i ${system_temp} shared.database.url ${DB_URL}
${YQ} w -i ${system_temp} shared.database.username ${DB_USER}
${YQ} w -i ${system_temp} shared.database.password ${DB_PASSWORD}

echo "Moving system file to proper location" | tee ${startup_log}
mv ${system_temp} /var/opt/jfrog/artifactory/etc/

cat /var/opt/jfrog/artifactory/etc/system.yaml >> ${startup_log}
echo "Creating Master Key" | tee ${startup_log}
echo ${ARTIFACTORY_MASTER_KEY} > /var/opt/jfrog/artifactory/etc/security/master.key

echo "Creating required directories for Artifactory configuration" | tee ${startup_log}
mkdir -p /var/opt/jfrog/artifactory/etc/artifactory

echo "Updating binarystore.xml with correct parameters" | tee ${startup_log}
sed -i -e "s/{{ s3_access_key }}/${S3_ACCESS_KEY}/g" ${binarystore_temp}
sed -i -e "s#{{ s3_access_secret_key }}#${S3_ACCESS_SECRET_KEY}#g" ${binarystore_temp}
sed -i -e "s/{{ s3_region }}/${S3_REGION}/g" ${binarystore_temp}
sed -i -e "s/{{ s3_bucket }}/${S3_BUCKET}/g" ${binarystore_temp}
mv ${binarystore_temp} /var/opt/jfrog/artifactory/etc/artifactory
