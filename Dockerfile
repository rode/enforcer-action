FROM golang:1.16-alpine as builder

WORKDIR /workspace

COPY go.mod go.sum /workspace/
RUN go mod download

COPY main.go main.go
COPY config/ config/
COPY action/ action/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o evaluate-resource-action

# ---------------
FROM gcr.io/distroless/static:latest

LABEL org.opencontainers.image.source=https://github.com/rode/evaluate-resource-action

COPY --from=builder /workspace/evaluate-resource-action /usr/local/bin/evaluate-resource-action

ENTRYPOINT ["evaluate-resource-action"]
