FROM golang:1.16 as builder

WORKDIR /go/src/github.com/shteou/helm-dependency-fetch

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY cmd cmd/
COPY pkg pkg/

RUN go build -o helm-dependency-fetch -ldflags="-w -s" cmd/fetch/main.go

FROM debian:buster-slim as production

COPY --from=builder /go/src/github.com/shteou/helm-dependency-fetch/helm-dependency-fetch /usr/bin/helm-dependency-fetch

CMD ["helm-dependency-fetch"]
