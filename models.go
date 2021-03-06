package fbvideo

import (
	"bytes"
	"encoding/json"
	"io"
)

type Map map[string]interface{}

type SessionInfo struct {
	*ChunkOffset
	UploadSessionID string `json:"upload_session_id"`
	VideoID         string `json:"video_id"`
}

type ChunkOffset struct {
	StartOffset string `json:"start_offset"`
	EndOffset   string `json:"end_offset"`
}

type ChunkInfo struct {
	Path        string
	Body        io.Reader
	ContentType string
}

type Result struct {
	Success bool `json:"success"`
}

type Error struct {
	Struct struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         int    `json:"190"`
		ErrorSubCode int    `json:"463"`
		FBTraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

func NewErrorFromBody(body io.ReadCloser) *Error {
	var err Error
	json.NewDecoder(body).Decode(&err)
	return &err
}

type LongLivedToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func NewLongLivedTokenFromBody(body io.ReadCloser) *LongLivedToken {
	var longLivedToken LongLivedToken
	json.NewDecoder(body).Decode(&longLivedToken)
	return &longLivedToken
}

func StringFromBody(body io.Reader) string {
	buf := &bytes.Buffer{}
	io.Copy(buf, body)
	return buf.String()
}
