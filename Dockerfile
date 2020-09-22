FROM golang:1.15-alpine as builder
LABEL maintainer=guillaume@villena.me

WORKDIR /ups-promtheus-exporter

COPY . .

RUN apk add --no-cache git
RUN go get
RUN go build

# -----

FROM alpine
COPY --from=builder /ups-promtheus-exporter/ups-promtheus-exporter /bin/ups-promtheus-exporter

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache nut
RUN ln -sf /usr/bin/upsc /bin/upsc

ENTRYPOINT ["/bin/ups-promtheus-exporter"]