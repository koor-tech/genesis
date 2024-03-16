#!/usr/bin/env sh

CONFIG_FILE="config.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Making config.yaml..."
    cat << EOF > $CONFIG_FILE
listen: ":8000"
mode: "debug"
directories:
  dataDir: "/koor"
database:
  host: "${GENESIS_DATABASE_HOST}"
  user: "${GENESIS_DATABASE_USER}"
  password: "${GENESIS_DATABASE_PASSWORD}"
  name: "${GENESIS_DATABASE_NAME}"
  sslEnabled: false
rabbitmq:
  host: "${GENESIS_RABBITMQ_HOST}"
  port: "${GENESIS_RABBITMQ_PORT}"
  user: "${GENESIS_RABBITMQ_USER}"
  password: "${GENESIS_RABBITMQ_PASSWORD}"
notifications:
  type: "noop"
  email:
    token: "${EMAIL_TOKEN}"
    from: "${EMAIL_FROM}"
    subject: "${EMAIL_SUBJECT}"
    replyTo: "${EMAIL_REPLYTO}"
cloudProvider:
  hetzner:
    token: "${GENESIS_HETZNER_TOKEN}"
EOF
else
    echo "the config.yaml exists skipping..."
fi

exec "$@"
