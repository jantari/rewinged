# syntax=docker/dockerfile:1

# Builder image to compile the binary. The tag has to satisfy the `go` directive
# in go.mod: the golang images set GOTOOLCHAIN=local, so an older builder fails
# outright rather than fetching the toolchain go.mod asks for.
FROM golang:1.25 AS builder

# Version metadata the binary reports through `-version`. Left at the defaults
# main.go declares when unset, so a plain `docker build .` still works.
ARG VERSION=development-build
ARG COMMIT=unknown
ARG COMPILE_TIME=unknown

WORKDIR $GOPATH/src/rewinged/rewinged/
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /tmp/rewinged -ldflags "-s -w \
      -X 'main.version=${VERSION}' \
      -X 'main.commit=${COMMIT}' \
      -X 'main.compileTime=${COMPILE_TIME}' \
      -X 'main.releaseMode=true'"



FROM alpine:latest AS certs

RUN apk add --no-cache ca-certificates



FROM scratch

COPY <<EOF /etc/passwd
rewinged:x:10002:10002:rewinged:/:/rewinged
EOF

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER rewinged

# WORKDIR creates directories if they don't exist; with 755 and owned by the current USER
WORKDIR /packages
WORKDIR /installers
WORKDIR /

COPY --from=builder /tmp/rewinged /rewinged

ENTRYPOINT ["/rewinged"]

