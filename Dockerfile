# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23 AS build-stage

WORKDIR /go/src

COPY go.mod go.sum /go/src/
RUN go mod download

COPY . /go/src

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /go/bin/cloudflare-ddns \
    cmd/main.go

# Deploy the application binary into a lean image
FROM scratch

ARG UID=1000
ARG GID=1000

USER ${UID}:${GID}

COPY --from=build-stage \
     /etc/ssl/certs/ca-certificates.crt \
     /etc/ssl/certs/

COPY --from=build-stage \
     --chown=${UID}:${GID} \
     /go/bin/cloudflare-ddns \
     /usr/local/bin/cloudflare-ddns

ENTRYPOINT ["cloudflare-ddns", "--repeat"]