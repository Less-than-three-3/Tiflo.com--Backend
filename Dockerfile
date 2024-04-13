FROM golang:1.21 AS builder
WORKDIR /app
COPY  . .
RUN go build -o tiflo_api cmd/main.go

FROM ubuntu:23.04 as run_stage
WORKDIR /out
COPY --from=builder /app/tiflo_api ./binary
RUN apt install ffmpeg
CMD ["./binary", "-python=false"]