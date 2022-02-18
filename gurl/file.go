package gurl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"sync"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto" //nolint:staticcheck // required for casting of V1 messages
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// FileResponse describes a GRPC response.
type FileResponse struct {
	Headers  metadata.MD
	Message  proto.Message
	Trailers metadata.MD
	Status   *status.Status
	Data     io.ReadCloser
}

// FileName returns the name of file received from response headers.
func (f *FileResponse) FileName() string {
	hvals := f.Headers.Get("Content-Disposition")
	if len(hvals) == 0 {
		return ""
	}
	_, params, err := mime.ParseMediaType(hvals[0])
	if err != nil {
		log.Printf("[ERROR] failed to parse mime of file: %v", err)
	}
	return params["filename"]
}

type fileRespData struct {
	*FileResponse
	lock *sync.Mutex
	err  error
}

func (*fileRespData) OnResolveMethod(md *desc.MethodDescriptor) {
	if txt, err := grpcurl.GetDescriptorText(md, nil); err == nil {
		log.Printf("[DEBUG] resolved method descriptor: %s", txt)
	}
}

func (*fileRespData) OnSendHeaders(md metadata.MD) {
	log.Printf("[DEBUG] sent headers: %v", md)
}

func (f *fileRespData) OnReceiveHeaders(md metadata.MD) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.Headers = md
}

func (f *fileRespData) OnReceiveResponse(message proto.Message) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.Message = message
	msg, ok := message.(*dynamic.Message)
	if !ok {
		f.err = fmt.Errorf("message is not of type *dynamic.Message: %w", ErrCast)
		return
	}

	body := httpbody.HttpBody{}
	if err := msg.ConvertTo(&body); err != nil {
		f.err = fmt.Errorf("convert message to *httpbody.Httpbody: %w", err)
		return
	}

	f.Data = io.NopCloser(bytes.NewReader(body.GetData()))
	log.Printf("[DEBUG] received response: %s, %v", body.GetContentType(), body.GetExtensions())
}

func (f *fileRespData) OnReceiveTrailers(s *status.Status, md metadata.MD) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.Status = s
	f.Trailers = md
	log.Printf("[DEBUG] received trailers: %v, %v", s, md)
}
