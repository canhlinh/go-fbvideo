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

	uploadSession := NewUploadSession("./testdata/SampleVideo_720x480_10mb.mp4", meID, accessToken)
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

	assert.Equal(t, uploadSession.fileName, "SampleVideo_720x480_10mb")
	assert.Equal(t, uploadSession.fileExt, ".mp4")
	assert.Equal(t, uploadSession.FilePath, "./testdata/SampleVideo_720x480_10mb.mp4")
	assert.NotNil(t, uploadSession.file)
	assert.EqualValues(t, uploadSession.fileSize, buf.Len())
	assert.Equal(t, uploadSession.fileChunkNumber, uint(0))
	assert.Equal(t, uploadSession.fileChunkFolder, "")
	assert.Equal(t, uploadSession.AccessToken, GetAccessToken())
	assert.NotNil(t, uploadSession.Transport)
	assert.Equal(t, uploadSession.Endpoint, fmt.Sprintf("https://graph-video.facebook.com/v2.6/%s/videos", meID))
	assert.EqualValues(t, uploadSession.ID, meID)
}

func TestUpload(t *testing.T) {

	accessToken := GetAccessToken()
	meID := GetMeID(accessToken)

	uploadSession := NewUploadSession("./testdata/SampleVideo_720x480_10mb.mp4", meID, accessToken)
	t.Log(uploadSession.fileSize)
	t.Run("UpdateWithPrivacySelf", func(t *testing.T) {
		videoID, err := uploadSession.Upload(Option{Privacy: &Privacy{Value: PrivacySelf}})
		if err != nil {
			t.Fatal(err)
		}

		videoInfo, err := GetResourceInfo(videoID, accessToken)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(videoInfo["source"], videoInfo["id"])
	})
}
