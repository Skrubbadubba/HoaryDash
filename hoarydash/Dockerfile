ARG BUILD_FROM
FROM ${BUILD_FROM:-golang:1.25-alpine} AS builder
WORKDIR /build
COPY server/ .
RUN go mod tidy && go build -o server .

FROM ${BUILD_FROM:-alpine:latest}
WORKDIR /app
COPY --from=builder /build/server .
COPY frontend/ ./frontend/
EXPOSE 4567
CMD ["./server"]