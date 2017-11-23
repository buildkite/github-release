FROM golang:1.9.2 AS build-env

WORKDIR /go/src/github.com/buildkite/github-release
ADD . /go/src/github.com/buildkite/github-release
RUN go build -o /bin/github-release

FROM alpine
WORKDIR /bin
COPY --from=build-env /bin/github-release /bin/
ENTRYPOINT /bin/github-release
