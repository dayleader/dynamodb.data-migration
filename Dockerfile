FROM golang:1.16-alpine as builder

# Compile application
WORKDIR /go/src/app
ADD go.mod /go/src/app/
RUN go mod download
ADD . /go/src/app
RUN CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.AppVersion=$(bash version.sh print_major_minor_patch)" -o main cmd/main.go

# Use a multi-stage build to reduce the size of the docker image.
FROM alpine:3

# Install packages
RUN apk add --no-cache bash

COPY --from=builder /go/src/app/main /bin/main
COPY --from=builder /go/src/app/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
