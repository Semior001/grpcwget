package main

import (
	"fmt"
	"log"
	"os"

	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"

	"github.com/Semior001/grpcwget/app/gurl"
	"github.com/google/uuid"
	"github.com/hashicorp/logutils"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	ProtoSets          []string      `short:"p" long:"protoset" required:"true" description:"location to pb file"`
	Headers            []string      `short:"H" long:"header" description:"headers to add to request"`
	RequestBody        string        `short:"d" long:"body" description:"body of GRPC request"`
	Addr               string        `short:"a" long:"addr" required:"true" description:"address to GRPC server"`
	Method             string        `short:"m" long:"method" required:"true" description:"full path to method"`
	OutputFileLocation string        `short:"o" long:"output" default:"." description:"location to output file, current dir by default"`
	Timeout            time.Duration `long:"timeout" description:"request timeout"`
	Debug              bool          `long:"dbg" description:"turn on debug mode"`
}

var version = "unknown"

func main() {
	fmt.Printf("grpcwget, version: %s\n", version)

	if _, err := flags.Parse(&opts); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	setupLog(opts.Debug)

	ctx := context.Background()
	if opts.Timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	if err := sendRequest(ctx); err != nil {
		log.Fatalf("failed to send request: %v", err)
	}
}

func setupLog(dbg bool) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: "INFO",
		Writer:   os.Stdout,
	}

	logFlags := log.Ldate | log.Ltime

	if dbg {
		logFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile
		filter.MinLevel = "DEBUG"
	}

	log.SetFlags(logFlags)
	log.SetOutput(filter)
}

func sendRequest(ctx context.Context) error {
	cl, err := gurl.NewClient(ctx, gurl.Params{
		Addr:          opts.Addr,
		Insecure:      true,
		ProtoSetPaths: opts.ProtoSets,
	})
	if err != nil {
		return fmt.Errorf("make client: %w", err)
	}

	resp, err := cl.GetFile(ctx, &gurl.Request{
		MethodURI: opts.Method,
		Headers:   opts.Headers,
		JSONBody:  io.NopCloser(strings.NewReader(opts.RequestBody)),
	})
	if err != nil {
		return fmt.Errorf("send request to get file: %w", err)
	}

	var f *os.File
	fInfo, err := os.Stat(opts.OutputFileLocation)
	if errors.Is(err, os.ErrNotExist) {
		if f, err = os.Create(opts.OutputFileLocation); err != nil {
			return fmt.Errorf("create file at %q: %w", opts.OutputFileLocation, err)
		}
		if fInfo, err = os.Stat(opts.OutputFileLocation); err != nil {
			return fmt.Errorf("get stat for %q file: %w", opts.OutputFileLocation, err)
		}
	}
	if err != nil {
		return fmt.Errorf("get output file location info: %w", err)
	}

	if fInfo.IsDir() {
		fName := resp.FileName()
		if fName == "" {
			fName = uuid.New().String()
		}

		fPath := path.Join(opts.OutputFileLocation, fName)
		if f, err = os.Create(fPath); err != nil {
			return fmt.Errorf("create file at %q: %w", fPath, err)
		}
	}

	if _, err = io.Copy(f, resp.Data); err != nil {
		return fmt.Errorf("write response file: %w", err)
	}

	return nil
}
