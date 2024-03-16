FROM golang:1.21 AS builder
WORKDIR /app
COPY  . .
RUN go build -o proxy cmd/main.go

FROM ubuntu:23.04 as run_stage
WORKDIR /out
COPY --from=builder /app/proxy ./binary
CMD ["./binary"]