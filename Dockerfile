FROM golang:1.21 AS builder
WORKDIR /app
COPY  . .
RUN go build -o tiflo_api cmd/main.go

FROM ubuntu:24.04 AS run_stage
WORKDIR /out
COPY --from=builder /app/tiflo_api ./binary
RUN apt update && \
    apt install -y ffmpeg && \
    apt clean
CMD ["./binary"]