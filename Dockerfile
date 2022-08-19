# builder image to compile the binary
FROM golang:1.16 as builder

WORKDIR $GOPATH/src/rewinged/rewinged/
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /tmp/rewinged



# final image with nothing but the binary
FROM scratch

COPY --from=builder /tmp/rewinged /rewinged

ENTRYPOINT ["/rewinged"]

