ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker cmd/worker/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrations/main.go



FROM alpine:latest

ENV HCLOUD_TOKEN token

RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl && \
    apk add --no-cache sudo && \
    apk add --no-cache openssh-client && \
    apk add --no-cache ca-certificates

RUN addgroup -S appgroup && adduser -S koor -G appgroup

USER koor

WORKDIR /home/koor/

## this should be a volume from ceph

COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=builder /usr/local/bin/kubeone /usr/local/bin/kubeone
COPY --from=builder /app/main .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrate .


USER root
RUN mkdir -p /koor/clients/templates
RUN chown koor:appgroup /home/koor/main
COPY templates/hetzner /koor/clients/templates/hetzner
COPY cmd/migrations/migrations /app/migrations
RUN chown -R koor:appgroup /koor/clients /home/koor/main /usr/bin/terraform /usr/local/bin/kubectl
USER koor

EXPOSE 8000

ENTRYPOINT ["./main"]