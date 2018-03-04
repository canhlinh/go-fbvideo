package fbvideo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	AccessToken = "EAADGAjjQ7sgBANGPH6KB2fsEZAf7EjCXJR0uoaVxQr25ODc4ZAeyrAoehV3ZAEOz9mrOKD5SxyJbVK7OwgsLeFhoM24c638f91zgNEnQ3ApGnN3OPe3t6S0ZAX3NwlTd3OC0sTZBKbUiw0TOz76l2jbEXwrZArb5qbDwUDfP0ZAsH1fxSv4BE4BRWZCPEzYgvzv8DRfGJgRts2tievuZATy6J"
)

func TestNewUploadSession(t *testing.T) {
	uploadSession := NewUploadSession("./testdata/video.mp4", 1849649771774166, AccessToken)
	if uploadSession == nil {
		t.Fatal("uploadSession should not be nill")
	}

	file, err := os.Open(uploadSession.FilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	buf := bytes.Buffer{}
	io.Copy(&buf, file)

	assert.Equal(t, uploadSession.fileName, "video")
	assert.Equal(t, uploadSession.fileExt, ".mp4")
	assert.Equal(t, uploadSession.FilePath, "./testdata/video.mp4")
	assert.Equal(t, uploadSession.fileData, buf.Bytes())
	assert.Equal(t, uploadSession.fileSize, buf.Len())
	assert.Equal(t, uploadSession.fileChunkNumber, 0)
	assert.Equal(t, uploadSession.fileChunkFolder, "")
	assert.Equal(t, uploadSession.AccessToken, AccessToken)
	assert.NotNil(t, uploadSession.Client)
	assert.Equal(t, uploadSession.Endpoint, fmt.Sprintf("https://graph-video.facebook.com/v2.3/%d/videos", 1849649771774166))
	assert.EqualValues(t, uploadSession.ID, 1849649771774166)
}

func TestUpload(t *testing.T) {
	uploadSession := NewUploadSession("./testdata/video.mp4", 1849649771774166, AccessToken)
	if err := uploadSession.Upload(); err != nil {
		t.Fatal(err)
	}
}
