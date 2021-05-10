FROM golang:1.16-alpine as builder

WORKDIR /workspace

COPY go.mod go.sum /workspace/
RUN go mod download

COPY main.go main.go
COPY config/ config/
COPY action/ action/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o action

# ---------------
FROM gcr.io/distroless/static:latest

LABEL org.opencontainers.image.source=https://github.com/rode/evaluate-policy-action

COPY --from=builder /workspace/action /usr/local/bin/action

ENTRYPOINT ["action"]
