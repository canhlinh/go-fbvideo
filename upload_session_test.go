package fbvideo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetAccessToken() string {
	return os.Getenv("FB_ACCESS_TOKEN")
}

func GetFbResourceID() int64 {
	s := os.Getenv("FB_RESOURCE_ID")
	resourceID, _ := strconv.ParseInt(s, 10, 64)
	return resourceID
}

func TestNewUploadSession(t *testing.T) {
	uploadSession := NewUploadSession("./testdata/video.mp4", GetFbResourceID(), GetAccessToken())
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
	assert.Equal(t, uploadSession.AccessToken, GetAccessToken())
	assert.NotNil(t, uploadSession.Client)
	assert.Equal(t, uploadSession.Endpoint, fmt.Sprintf("https://graph-video.facebook.com/v2.6/%d/videos", GetFbResourceID()))
	assert.EqualValues(t, uploadSession.ID, GetFbResourceID())
}

func TestUpload(t *testing.T) {
	uploadSession := NewUploadSession("./testdata/video.mp4", GetFbResourceID(), GetAccessToken())

	t.Run("UpdateWithDefaultPrivacy", func(t *testing.T) {
		if err := uploadSession.Upload(Option{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("UpdateWithPrivacySelf", func(t *testing.T) {
		if err := uploadSession.Upload(Option{Privacy: &Privacy{Value: PrivacySelf}}); err != nil {
			t.Fatal(err)
		}
	})
}
