FROM golang:1.20 AS builder
WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download

COPY . /build
ENV CGO_ENABLED=0
RUN go build
RUN groupadd -g 999 user && \
    useradd -r -u 999 -g user user
FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
USER user
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/jq-check-with-body /app/jq-check-with-body
ENTRYPOINT ["/app/jq-check-with-body"]
