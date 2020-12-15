FROM golang:1.15 as builder

WORKDIR /go/src/github.com/shteou/helm-dependency-fetch

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY main.go main.go
COPY handlers handlers
COPY k8s k8s

RUN go build -ldflags="-w -s" main.go
RUN mv main helm-dependency-fetch

FROM debian:buster-slim as production

COPY --from=builder /go/src/github.com/shteou/helm-dependency-fetch/helm-dependency-fetch /usr/bin/helm-dependency-fetch

CMD ["helm-dependency-fetch"]
