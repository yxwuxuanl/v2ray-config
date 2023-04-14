FROM --platform=linux/amd64 golang:1.20

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o v2ray-config && chmod +x ./v2ray-config


FROM --platform=linux/amd64 alpine:3.15.5

COPY --from=0 /app/v2ray-config /v2ray-config

ENTRYPOINT ["/v2ray-config"]