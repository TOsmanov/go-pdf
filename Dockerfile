FROM golang:1.21 AS builder
COPY ./ /usr/local/go/src/go_pdf/
WORKDIR /usr/local/go/src/go_pdf
RUN go clean --modcache && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -mod=readonly -o gopdf /usr/local/go/src/go_pdf/server/main.go

FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app

# Installing required packages
RUN apt-get -y update; \
    apt-get install -y wget

# Chromium install
RUN apt-get install -y software-properties-common \
    && add-apt-repository -y ppa:xtradeb/apps
RUN apt update \
    && apt-cache policy chromium \
    && apt install -y chromium \
    && apt-cache policy chromium

# Copy binary file from builder
COPY --from=builder /usr/local/go/src/go_pdf/gopdf /app/
CMD ["/app/gopdf"]
