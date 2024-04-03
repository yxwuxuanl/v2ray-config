ARG V2RAY_TAG

FROM golang:1.22.0

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build .

FROM teddysun/v2ray:${V2RAY_TAG}

COPY --from=0 /app/v2ray-subscribe /v2ray-subscribe

ENTRYPOINT ["/v2ray-subscribe"]