# grpcwget [![Go](https://github.com/Semior001/grpcwget/actions/workflows/.go.yaml/badge.svg)](https://github.com/Semior001/grpcwget/actions/workflows/.go.yaml) [![codecov](https://codecov.io/gh/Semior001/grpcwget/branch/master/graph/badge.svg?token=nLxLt9Vdyo)](https://codecov.io/gh/Semior001/grpcwget) [![go report card](https://goreportcard.com/badge/github.com/Semior001/grpcwget)](https://goreportcard.com/report/github.com/Semior001/grpcwget) [![Go Reference](https://pkg.go.dev/badge/github.com/Semior001/grpcwget.svg)](https://pkg.go.dev/github.com/Semior001/grpcwget)
Small utility to download files through grpc. Currently, only files received via 
[google.api.HttpBody](https://github.com/googleapis/googleapis/blob/master/google/api/httpbody.proto) 
are supported.

### install
`go install github.com/Semior001/grpcwget@v0.1.1` or via downloading the binary
in the releases.

## options
```
Application Options:
  -p, --protoset= location to pb file
  -H, --header=   headers to add to request
  -d, --body=     body of GRPC request
  -a, --addr=     address to GRPC server
  -m, --method=   full path to method
  -o, --output=   location to output file, current dir by default (default: .)
      --timeout=  request timeout
      --dbg       turn on debug mode
```

### example
```bash
grpcwget -p api.pb -a localhost:9000 \
      -m semior.some.package.v1.Service/DownloadFile \
      -d '{"someparam": "someval"}' \
      -o some_response.pdf
```
