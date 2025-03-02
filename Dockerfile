FROM golang:1.24-alpine as build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o wordpress-go-proxy ./cmd/server

FROM scratch
COPY --from=build /build/wordpress-go-proxy /wordpress-go-proxy
COPY --from=build /build/templates /templates
COPY --from=build /build/static /static

ENTRYPOINT ["/wordpress-go-proxy"]