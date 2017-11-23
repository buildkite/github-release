FROM golang:1.9.2 AS build-env

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/github.com/buildkite/github-release
ADD . /go/src/github.com/buildkite/github-release
RUN go build -a -tags netgo -ldflags '-w' -o /bin/github-release

FROM scratch
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env /bin/github-release /github-release
ENTRYPOINT ["/github-release"]
