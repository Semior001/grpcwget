package gurl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck // required for grpcurl
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
)

const httpBodyName = "google.api.HttpBody"

// Client provides methods to make grpc requests to remote server.
type Client struct {
	conn *grpc.ClientConn
	ds   grpcurl.DescriptorSource
}

// Params describes parameters of grpc client.
type Params struct {
	Addr          string
	Insecure      bool
	ProtoSetPaths []string
}

// NewClient makes new instance of Client.
func NewClient(ctx context.Context, params Params) (*Client, error) {
	var dialOpts []grpc.DialOption

	if params.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, params.Addr, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ds, err := grpcurl.DescriptorSourceFromProtoSets(params.ProtoSetPaths...)
	if err != nil {
		return nil, fmt.Errorf("get descriptor source from protosets: %w", err)
	}

	return &Client{
		conn: conn,
		ds:   ds,
	}, nil
}

// Request describes a GRPC request.
type Request struct {
	MethodURI string
	Headers   []string // in form or "HeaderName: HeaderValue", see grpcurl.MetadataFromHeaders
	JSONBody  io.ReadCloser
}

// GetFile downloads file with the requested GRPC method.
// Requested GRPC method must have output of type google.api.HttpBody.
func (c *Client) GetFile(ctx context.Context, req *Request) (*FileResponse, error) {
	mtdDesc, err := c.locateMethod(req.MethodURI)
	if err != nil {
		return nil, fmt.Errorf("locate method %q in descriptor source: %w",
			req.MethodURI, err)
	}

	if outName := mtdDesc.GetOutputType().GetFullyQualifiedName(); outName != httpBodyName {
		return nil, ErrOutputNotSupported(outName)
	}

	respData := &fileRespData{
		FileResponse: &FileResponse{Data: nil},
		lock:         &sync.Mutex{},
	}

	err = grpcurl.InvokeRPC(ctx, c.ds, c.conn,
		req.MethodURI, req.Headers,
		respData, grpcurl.NewJSONRequestParserWithUnmarshaler(
			req.JSONBody,
			jsonpb.Unmarshaler{
				AnyResolver:        grpcurl.AnyResolverFromDescriptorSource(c.ds),
				AllowUnknownFields: false,
			},
		).Next,
	)
	if err != nil {
		return nil, fmt.Errorf("invoke rpc: %w", err)
	}

	if err = req.JSONBody.Close(); err != nil {
		return nil, fmt.Errorf("close request body: %w", err)
	}

	if respData.err != nil {
		return nil, respData.err
	}

	if respData.Data == nil {
		return nil, ErrNoResponse
	}

	return respData.FileResponse, nil
}

func (c *Client) locateMethod(methodURI string) (*desc.MethodDescriptor, error) {
	serviceName, methodName := parseMethodURI(methodURI)

	dsc, err := c.ds.FindSymbol(serviceName)
	if err != nil {
		return nil, fmt.Errorf("find method symbol: %w", err)
	}

	svcDesc, ok := dsc.(*desc.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("target server does not expose service %q: %w", serviceName, ErrCast)
	}

	mtdDesc := svcDesc.FindMethodByName(methodName)
	if mtdDesc == nil {
		return nil, ErrMethodNotFound(methodName)
	}

	return mtdDesc, nil
}

func parseMethodURI(methodURI string) (string, string) {
	pos := strings.LastIndex(methodURI, "/")
	if pos < 0 {
		pos = strings.LastIndex(methodURI, ".")
		if pos < 0 {
			return "", ""
		}
	}
	return methodURI[:pos], methodURI[pos+1:]
}
