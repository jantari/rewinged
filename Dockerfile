# syntax=docker/dockerfile:1

# Builder image to compile the binary
FROM golang:1.19 as builder

WORKDIR $GOPATH/src/rewinged/rewinged/
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /tmp/rewinged -ldflags '-X "main.releaseMode=true"'



# Final image with nothing but the binary
FROM scratch

COPY <<EOF /etc/passwd
rewinged:x:10002:10002:rewinged:/:/rewinged
EOF

# WORKDIR creates directories if they don't exist; with 755 and owned by root
WORKDIR /packages
WORKDIR /installers
WORKDIR /

COPY --from=builder /tmp/rewinged /rewinged

USER rewinged
ENTRYPOINT ["/rewinged"]

