ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

WORKDIR /app

RUN apk --no-cache add curl unzip


RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/

RUN curl -O https://releases.hashicorp.com/terraform/1.0.11/terraform_1.0.11_linux_amd64.zip && \
    unzip terraform_1.0.11_linux_amd64.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform_1.0.11_linux_amd64.zip

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

ENV HCLOUD_TOKEN token

RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl && \
    apk add --no-cache sudo && \
    apk add --no-cache ca-certificates

RUN addgroup -S appgroup && adduser -S koor -G appgroup

USER koor

WORKDIR /home/koor/

## this should be a volume from ceph

COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=builder /app/main .


USER root
RUN mkdir -p /koor/clients/templates
RUN chown koor:appgroup /home/koor/main
RUN curl -sfL get.kubeone.io | sh
RUN cp -r /home/koor/kubeone_1.7.2_linux_amd64/examples/terraform/* /koor/clients/templates/
RUN chown -R koor:appgroup /koor/clients /home/koor/main /usr/local/bin/terraform /usr/local/bin/kubectl
USER koor

EXPOSE 8000

ENTRYPOINT ["./main"]