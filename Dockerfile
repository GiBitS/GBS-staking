ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/-stakin-eus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The stakin-eus Authors <stakin-eus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/node_exporter /bin/node_exporter

EXPOSE      9100
USER        nobody
ENTRYPOINT  [ "/bin/node_exporter" ]
