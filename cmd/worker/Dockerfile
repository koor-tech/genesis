ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker cmd/worker/main.go

FROM scratch

WORKDIR /home/koor/

COPY --from=builder /app/worker .
ENTRYPOINT ["/home/koor/worker", "monitor"]