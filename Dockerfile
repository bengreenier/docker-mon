FROM golang:1.14.4-alpine3.11 as core
RUN apk update

FROM core as build-base
RUN apk add git jq gcc libc-dev
RUN GO111MODULE=on go get -u -v \
  github.com/mdempsky/gocode@latest \
  github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest \
  github.com/ramya-rao-a/go-outline@latest \
  github.com/rogpeppe/godef@latest \
  github.com/sqs/goreturns@latest \
  golang.org/x/lint/golint@latest \
  golang.org/x/tools/gopls@latest \
  github.com/golang/mock/mockgen@latest

RUN go get -v github.com/go-delve/delve/cmd/dlv

RUN go get -v -d \
  github.com/stamblerre/gocode

RUN go build -o gocode-gomod github.com/stamblerre/gocode && mv gocode-gomod $GOPATH/bin

FROM build-base as builder
WORKDIR /go/src/github.com/bengreenier/docker-mon
COPY . .

RUN go get -d -v ./...
RUN go build -o ./bin/mon ./cmd/mon

# If you wanted to rebuild mocks each time we built, you could uncomment this
#
# RUN cd internal/app/mon && mockgen -destination ./mocks/dockerd.go . DockerAPI

RUN go test -cover ./...

FROM core as runner
COPY --from=builder /go/src/github.com/bengreenier/docker-mon/bin/mon /usr/bin/mon

LABEL mon.ignore=1
VOLUME ["/var/run/docker.sock"]
CMD ["mon"]
