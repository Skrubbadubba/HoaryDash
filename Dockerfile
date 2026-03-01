FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY server/ .
RUN go build -o server .

# ----

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/server .
COPY frontend/static ./frontend/static

EXPOSE 4567
CMD ["./server"]