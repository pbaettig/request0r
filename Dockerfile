FROM golang:1.11 as builder
RUN mkdir -p /go/src/github.com/pbaettig/request0r/cmd/rq0r && \
go get gopkg.in/yaml.v2 github.com/sirupsen/logrus
COPY . /go/src/github.com/pbaettig/request0r
WORKDIR /go/src/github.com/pbaettig/request0r/cmd/rq0r
RUN go test github.com/pbaettig/request0r/internal/app && \
    go test github.com/pbaettig/request0r/pkg/randurl && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o rq0r main.go


FROM scratch
COPY --from=builder /go/src/github.com/pbaettig/request0r/cmd/rq0r/rq0r /app/
WORKDIR /app
ENTRYPOINT ["/app/rq0r"]