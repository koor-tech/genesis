FROM golang:alpine as genesis-base-builder

WORKDIR /app

RUN apk --no-cache add curl unzip

COPY . .

RUN go mod download
RUN apk --no-cache add curl unzip


RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/

RUN curl -O https://releases.hashicorp.com/terraform/1.0.11/terraform_1.0.11_linux_amd64.zip && \
    unzip terraform_1.0.11_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.0.11_linux_amd64.zip

