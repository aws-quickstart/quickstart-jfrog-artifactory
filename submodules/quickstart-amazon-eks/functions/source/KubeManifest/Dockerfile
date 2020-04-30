FROM lambci/lambda:build-python3.7

COPY . .

RUN pip install -t . -r ./requirements.txt && \
    rm -rf *.dist-info *.pth && \
    rm Dockerfile requirements.txt && \
    zip -X -r lambda.zip ./

CMD mkdir -p /output/ && mv lambda.zip /output/
