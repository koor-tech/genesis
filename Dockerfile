ARG BUILDER=koor-tech/genesis-base-builder:latest
FROM $BUILDER as builder

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o genesis .

FROM docker.io/library/alpine:3.19.1

RUN apk add --no-cache curl sudo openssh-client ca-certificates && \
    addgroup -S appgroup && adduser -S koor -G appgroup

USER koor

WORKDIR /home/koor/

COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=builder /usr/local/bin/kubeone /usr/local/bin/kubeone
COPY --from=builder /app/genesis .

USER root
RUN chown koor:appgroup /home/koor/genesis
RUN chown -R koor:appgroup /home/koor/genesis /usr/local/bin/terraform /usr/local/bin/kubectl
COPY setup.sh /usr/local/bin/setup.sh
RUN chmod +x /usr/local/bin/setup.sh
USER koor

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/setup.sh"]
CMD ["./genesis"]
