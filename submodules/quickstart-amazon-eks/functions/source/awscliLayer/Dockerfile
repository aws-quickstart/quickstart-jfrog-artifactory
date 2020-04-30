FROM lambci/lambda:build-python3.7

COPY . .

RUN mkdir -p python/ && \
    pip install -t python/ -r ./requirements.txt && \
    rm -rf python/*.dist-info python/*.pth python/__pycache__ python/bin/__pycache__ && \
    mv python/bin ./ && \
    mv python/* ./bin/ && \
    rm Dockerfile requirements.txt && \
    zip -X -r lambda.zip ./

CMD mkdir -p /output/ && mv lambda.zip /output/
