ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

WORKDIR /app

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
COPY --from=builder /usr/local/bin/kubeone /usr/local/bin/kubeone
COPY --from=builder /app/main .


USER root
RUN mkdir -p /koor/clients/templates
RUN chown koor:appgroup /home/koor/main
COPY templates/hetzner /koor/clients/templates/hetzner
RUN chown -R koor:appgroup /koor/clients /home/koor/main /usr/local/bin/terraform /usr/local/bin/kubectl
USER koor

EXPOSE 8000

ENTRYPOINT ["./main"]