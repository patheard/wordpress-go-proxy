FROM golang:1.24-alpine as build

WORKDIR /build

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o wordpress-go-proxy ./cmd/server

FROM scratch
COPY --from=build /build/wordpress-go-proxy /wordpress-go-proxy
COPY --from=build /build/templates /templates
COPY --from=build /build/static /static

ENTRYPOINT ["/wordpress-go-proxy"]