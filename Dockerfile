FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o route-validator .

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/route-validator /route-validator
COPY certs /certs
ENTRYPOINT ["/route-validator"]