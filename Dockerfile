# Go build
FROM golang:1.21-alpine3.19 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o genesis .

# Kubectl
FROM docker.io/library/alpine:3.19 as kubectl

ENV KUBECTL_VERSION="1.29.2"

RUN apk --no-cache add curl && \
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/

# Terraform
FROM docker.io/library/alpine:3.19 as terraform

ENV TERRAFORM_VERSION="1.0.11"

RUN apk --no-cache add curl && \
    curl -O "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    mv terraform /usr/local/bin/ && \
    rm "terraform_${TERRAFORM_VERSION}_linux_amd64.zip"

# App Image
FROM docker.io/library/alpine:3.19

ENV KUBEONE_VERSION="1.7.2"

RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl && \
    apk add --no-cache sudo && \
    apk add --no-cache ca-certificates && \
    addgroup -S appgroup && \
    adduser -S koor -G appgroup && \
    mkdir -p /koor/clients/templates && \
    curl -sfL get.kubeone.io | sh && \
    cp -r "kubeone_${KUBEONE_VERSION}_linux_amd64/examples/terraform/"* /koor/clients/templates/ && \
    rm -rf "kubeone_${KUBEONE_VERSION}_linux_amd64" && \
    chown -R koor:appgroup /koor/clients

COPY --from=builder /app/genesis /usr/local/bin/genesis
COPY --from=kubectl /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=terraform /usr/local/bin/terraform /usr/local/bin/terraform

WORKDIR /koor/

USER koor

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/genesis"]
