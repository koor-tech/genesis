ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker cmd/worker/main.go

FROM alpine

RUN apk update && \
    apk --no-cache add openssh

WORKDIR /home/koor/

COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=builder /usr/local/bin/terraform /usr/bin/terraform
COPY --from=builder /app/worker .
COPY templates/hetzner/variables.tf  /koor/clients/templates/hetzner/variables.tf
RUN cat  /koor/clients/templates/hetzner/variables.tf

ENTRYPOINT ["/home/koor/worker", "monitor"]
