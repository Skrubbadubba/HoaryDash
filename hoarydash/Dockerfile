FROM node:22-alpine AS icons
WORKDIR /build
RUN npm install @mdi/js
COPY extract-icons.mjs .
RUN node extract-icons.mjs

FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY server/ .
COPY --from=icons /build/mdi.json ./mdi.json
RUN go mod download && go build -o server .

ARG BUILD_FROM
FROM ${BUILD_FROM:-alpine:latest}
WORKDIR /app
COPY --from=builder /build/server .
COPY frontend/ ./frontend/
EXPOSE 4567
CMD ["./server"]