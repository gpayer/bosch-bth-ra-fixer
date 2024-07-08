ARG BUILD_FROM
FROM $BUILD_FROM as base

FROM golang:1.22-alpine as builder

COPY . /app
RUN cd /app && go build -o /fixer .

FROM base as release

COPY --from=builder /fixer /usr/bin/fixer
COPY rootfs /
