FROM lambci/lambda:build-python3.7

ENV VERSION="2.16.3"

COPY . .

RUN curl -L https://get.helm.sh/helm-v${VERSION}-linux-amd64.tar.gz |tar xvz && \
    mv linux-amd64/helm bin/helm-client && \
    mv linux-amd64/tiller bin/tiller && \
    chmod +x bin/helm-client && \
    chmod +x bin/tiller && \
    rm -rf linux-amd64 Dockerfile && \
    zip -X -r lambda.zip ./

CMD mkdir -p /output/ && mv lambda.zip /output/
