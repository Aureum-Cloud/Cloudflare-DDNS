# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23 AS build-stage

WORKDIR /go/src

COPY go.mod go.sum /go/src/
RUN go mod download

COPY . /go/src

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/cloudflare-ddns bootstrap/main.go

# Deploy the application binary into a lean image
FROM alpine:3.19 AS build-release-stage

COPY --from=build-stage /go/bin/cloudflare-ddns /usr/local/bin/cloudflare-ddns

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot:nonroot

ENTRYPOINT ["cloudflare-ddns", "--repeat"]