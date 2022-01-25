FROM golang:1.12.4
WORKDIR /go/src/github.com/paralin/replay-spreadsheet/
ADD ./ ./
RUN mkdir -p bin/ && \
  GO111MODULE=on CGO_ENABLED=0 GOOS=linux go \
  build -v -a \
  -o ./bin/replay-spreadsheet ./cmd/replay-spreadsheet

FROM alpine:latest
ENV GODEBUG netdns=cgo
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/paralin/replay-spreadsheet/bin/replay-spreadsheet ./
CMD ["/root/replay-spreadsheet"]
