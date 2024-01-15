package convert

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	errors "golang.org/x/xerrors"

	"github.com/abeja-inc/abeja-platform-model-proxy/config"
	"github.com/abeja-inc/abeja-platform-model-proxy/entity"
)

// DummyMethodForResponse is dummy-http-method for response.
const DummyMethodForResponse = "response"

// KeyContentType is response header key of Content-Type
const KeyContentType = "Content-Type"

// KeyAbejaProxyVersion is response header key of x-abeja-sys-meta-proxy-version
const KeyAbejaProxyVersion = "X-Abeja-Sys-Meta-Proxy-Version"

// KeyContentLength is response header key of Content-Length
const KeyContentLength = "Content-Length"

// KeyConnection is response header key of Connection
const KeyConnection = "Connection"

// ConverterError is custom error struct.
type ConverterError struct {
	Msg        string
	StatusCode int
	Err        error
	frame      errors.Frame
}

func (err *ConverterError) Error() string {
	return err.Msg
}

func (err *ConverterError) Unwrap() error {
	return err.Err
}

func (err *ConverterError) Format(s fmt.State, v rune) {
	errors.FormatError(err, s, v)
}

func (err *ConverterError) FormatError(p errors.Printer) error {
	p.Print(err.Error())
	err.frame.Format(p)
	return err.Err
}

type converter interface {
	IsTarget(ctx context.Context, method string, contentType string) bool
	ToContent(ctx context.Context, r *http.Request, conf *config.Configuration) (*entity.ContentList, error)
	FromResponse(ctx context.Context, res entity.Response) (int, map[string]string, *os.File, error)
}

var converters []converter

func init() {
	converters = append(converters, &defaultConverter{})
	converters = append(converters, &multipartConverter{})
}

// ToContents returns ContentList that converted from Request.
func ToContents(
	ctx context.Context,
	r *http.Request,
	conf *config.Configuration) (*entity.ContentList, error) {

	var targetConverter converter
	contentType := r.Header.Get("Content-Type")
	for _, conv := range converters {
		if conv.IsTarget(ctx, r.Method, contentType) {
			targetConverter = conv
			break
		}
	}
	if targetConverter == nil {
		return nil, &ConverterError{
			Msg:        fmt.Sprintf("Content-Type: [%s] is not supported", contentType),
			StatusCode: http.StatusNotImplemented,
			Err:        nil,
			frame:      errors.Caller(0),
		}
	}

	cl, err := targetConverter.ToContent(ctx, r, conf)
	if err != nil {
		return cl, err
	}

	var headers []*entity.Header
	for key, value := range r.Header {
		header := &entity.Header{
			Key:    strings.ToLower(key),
			Values: value,
		}
		headers = append(headers, header)
	}
	sort.Slice(headers, func(i, j int) bool { return headers[i].Key < headers[j].Key })
	cl.Headers = headers
	return cl, nil
}

// FromResponse returns information of http response from struct of Response.
func FromResponse(ctx context.Context, res entity.Response) (int, map[string]string, *os.File, error) {
	var targetConverter converter
	contentType := res.ContentType
	if contentType == nil {
		targetConverter = &defaultConverter{}
		return targetConverter.FromResponse(ctx, res)
	}

	for _, conv := range converters {
		if conv.IsTarget(ctx, DummyMethodForResponse, *contentType) {
			targetConverter = conv
			break
		}
	}
	if targetConverter == nil {
		return 0, nil, nil, &ConverterError{
			Msg:        fmt.Sprintf("Content-Type: [%s] is not supported", *contentType),
			StatusCode: http.StatusNotImplemented,
			Err:        nil,
			frame:      errors.Caller(0),
		}
	}

	return targetConverter.FromResponse(ctx, res)
}
