FROM golang:1.18 as builder
 
WORKDIR /go/src/impression
COPY . .
RUN go mod download && CGO_ENABLED=0 make bin/impression bin/impression.test

# This image is the release build
FROM gcr.io/distroless/static-debian11 AS release
COPY --from=builder /go/src/impression/bin/impression /

CMD ["/impression"]

# This image is the test build
FROM gcr.io/distroless/static-debian11 AS test
COPY --from=builder /go/src/impression/bin/impression.test /

CMD ["/impression.test"]