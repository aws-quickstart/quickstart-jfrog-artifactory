#/bin/bash

source functions.sh

logger $(whoami)

NGINX_CONF=/tmp/artifactory.conf
SSL_DIR=/var/opt/jfrog/nginx/ssl

logger "Copying SSL certs"
echo ${CERTIFICATE} | base64 --decode > $SSL_DIR/cert.pem
echo ${CERTIFICATE_KEY} | base64 --decode > $SSL_DIR/cert.key

logger "Creating artifactory.conf"
sed -i -e "s#{{ ssl_dir }}#$SSL_DIR#g" $NGINX_CONF
sed -i -e "s#{{ key_dir }}#$SSL_DIR#g" $NGINX_CONF
sed -i -e "s/{{ artifactory_server_name }}/${JCR_SERVER_NAME}/g" $NGINX_CONF
sed -i -e "s/{{ certificate_domain }}/${CERTIFICATE_DOMAIN}/g" $NGINX_CONF
sed -i '/elif artifactory_major_verion/,$d' $NGINX_CONF
echo "}" >> $NGINX_CONF
cat $NGINX_CONF | grep -v -e "{% if ecs_deployment %}" -e "{% else %}" -e "{% endif %}" -e "server artifactory:808" -e "artifactory_major_verion" > /var/opt/jfrog/nginx/conf.d/artifactory.conf
