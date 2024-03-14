FROM docker.io/library/golang:1.21-alpine3.19 as genesis-base-builder

ENV KUBEONE_VERSION=1.7.2
ENV DOCKER_OS_NAME=linux_amd64
ENV KUBEONE_ZIP=kubeone_${KUBEONE_VERSION}_${DOCKER_OS_NAME}.zip
ENV KUBEONE_DIR=kubeone_${KUBEONE_VERSION}_${DOCKER_OS_NAME}

WORKDIR /app

RUN apk --no-cache add curl unzip

COPY . .

RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/

RUN curl -O https://releases.hashicorp.com/terraform/1.0.11/terraform_1.0.11_linux_amd64.zip && \
    unzip terraform_1.0.11_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.0.11_linux_amd64.zip

RUN  curl -LO https://github.com/kubermatic/kubeone/releases/download/v${KUBEONE_VERSION}/${KUBEONE_ZIP} && \
     unzip ${KUBEONE_ZIP}  -d ${KUBEONE_DIR} && \
     mv ${KUBEONE_DIR}/kubeone /usr/local/bin

RUN go mod download && go mod tidy
