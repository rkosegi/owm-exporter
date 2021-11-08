FROM golang:1.16 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download -x

COPY main.go main.go
COPY config/ config/
COPY client/ client/
COPY types/ types/
COPY exporter/ exporter/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o owm-exporter .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/owm-exporter /
USER 65532:65532

EXPOSE 9111

ENTRYPOINT ["/owm-exporter"]
