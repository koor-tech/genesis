ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrations/main.go


FROM alpine:latest
RUN apk add --no-cache bash

COPY --from=builder /app/migrate /genesis/bin/migrate
COPY cmd/migrations/migrations /genesis/bin/migrations
ENTRYPOINT ["/genesis/bin/migrate", "migrate", "up"]