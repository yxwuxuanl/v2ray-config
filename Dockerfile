FROM bcycr-registry.cn-hangzhou.cr.aliyuncs.com/mirror/golang:1.20

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o v2ray-config && chmod +x ./v2ray-config


FROM bcycr-registry.cn-hangzhou.cr.aliyuncs.com/mirror/alpine:3.15.5

COPY --from=0 /app/v2ray-config /v2ray-config

ENTRYPOINT ["/v2ray-config"]