FROM golang:1.15-alpine as builder
RUN set -ex; \
  go env -w GO111MODULE=on; \
  go env -w GOPROXY=https://goproxy.io,direct; \
  go get github.com/ztino/jd_seckill

FROM alpine:3.12

LABEL MAINTAINER="currycan/helloworld"

ENV TZ Asia/Shanghai

RUN set -ex; apk upgrade; apk add --no-cache --no-progress bash tzdata busybox-extras; \
        ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime; \
        echo ${TZ} > /etc/timezone; \
        rm -rf /var/cache/apk/*;

COPY --from=builder /go/bin/jd_seckill /usr/local/bin/jd_seckill

WORKDIR /app

CMD ["jd_seckill"]
