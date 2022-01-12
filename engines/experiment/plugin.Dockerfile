FROM alpine:3.15

ARG turing_user="turing"
ARG turing_user_group="app"

RUN addgroup -S ${turing_user_group} \
    && adduser -S ${turing_user} -G ${turing_user_group} -H

ARG PLUGIN_BINARY

ENV PLUGIN_NAME ""
ENV PLUGINS_DIR "/app/plugins"

ADD --chown=${turing_user}:${turing_user_group} ${PLUGIN_BINARY} /go/bin/plugin

CMD ["sh", "-c", "cp /go/bin/plugin ${PLUGINS_DIR}/${PLUGIN_NAME:?variable must be SET}"]
