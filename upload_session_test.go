package fbvideo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetMeID(accesstoken string) string {
	meInfo, err := GetMe(accesstoken)
	if err != nil {
		panic(err)
	}
	meID := meInfo["id"].(string)
	return meID
}

func TestNewUploadSession(t *testing.T) {

	accessToken := GetAccessToken()
	meID := GetMeID(accessToken)

	uploadSession := NewUploadSession("./testdata/video.mp4", meID, accessToken)
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
	assert.Equal(t, uploadSession.Endpoint, fmt.Sprintf("https://graph-video.facebook.com/v2.6/%s/videos", meID))
	assert.EqualValues(t, uploadSession.ID, meID)
}

func TestUpload(t *testing.T) {

	accessToken := GetAccessToken()
	meID := GetMeID(accessToken)

	uploadSession := NewUploadSession("./testdata/video.mp4", meID, accessToken)

	t.Run("UpdateWithDefaultPrivacy", func(t *testing.T) {
		if videoID, err := uploadSession.Upload(Option{}); err != nil {
			t.Fatal(err)
		} else {
			t.Log(videoID)
		}
	})

	t.Run("UpdateWithPrivacySelf", func(t *testing.T) {
		if _, err := uploadSession.Upload(Option{Privacy: &Privacy{Value: PrivacySelf}}); err != nil {
			t.Fatal(err)
		}
	})
}
