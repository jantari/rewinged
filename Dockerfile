# syntax=docker/dockerfile:1

# Builder image to compile the binary
FROM golang:1.21 as builder

WORKDIR $GOPATH/src/rewinged/rewinged/
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /tmp/rewinged -ldflags '-X "main.releaseMode=true"'



FROM alpine:latest as certs

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

