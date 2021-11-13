FROM golang:1.16 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY config.go config.go
COPY client.go client.go
COPY types.go types.go
COPY exporter.go exporter.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o owm-exporter .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/owm-exporter /
USER 65532:65532

EXPOSE 9111

ENTRYPOINT ["/owm-exporter"]
