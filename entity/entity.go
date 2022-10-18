package entity

import "context"

// Content is struct of part of HTTP-Request.
type Content struct {
	ContentType *string                `json:"content_type,omitempty"`
	Path        *string                `json:"path,omitempty"`
	FileName    *string                `json:"file_name,omitempty"`
	FormName    *string                `json:"form_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // metadata of the content
}

// Header is struct of part of headers of HTTP-Request
type Header struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

// ContentList is struct of HTTP-Request.
type ContentList struct {
	Method         string          `json:"method"`
	ContentType    string          `json:"content_type"`
	Headers        []*Header       `json:"headers"`
	Contents       []*Content      `json:"contents"`
	AsyncRequestID string          `json:"-"`
	AsyncARMSToken string          `json:"-"`
	Ctx            context.Context `json:"-"`
}

// Response is struct of HTTP-Response.
type Response struct {
	ContentType *string            `json:"content_type,omitempty"`
	Metadata    *map[string]string `json:"metadata,omitempty"`
	Path        *string            `json:"path,omitempty"`
	ErrMsg      *string            `json:"error_message,omitempty"`
	StatusCode  *int               `json:"status_code,omitempty"`
}
