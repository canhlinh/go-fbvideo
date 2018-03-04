package fbvideo

import (
	"encoding/json"
	"io"
)

type SessionInfo struct {
	*ChunkOffset
	UploadSessionID string `json:"upload_session_id"`
	VideoID         string `json:"video_id"`
}

type ChunkOffset struct {
	StartOffset string `json:"start_offset"`
	EndOffset   string `json:"end_offset"`
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
