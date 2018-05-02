package fbvideo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

//PrivacyValue Determines the privacy settings of the video.
// If not supplied, this defaults to the privacy level granted to the app in the Login Dialog.
// This field cannot be used to set a more open privacy setting than the one granted.
type PrivacyValue string

const (
	PrivacyEveryOne         PrivacyValue = "EVERYONE"
	PrivacyAllFriends       PrivacyValue = "ALL_FRIENDS"
	PrivacyFriendsOfFriends PrivacyValue = "FRIENDS_OF_FRIENDS"
	PrivacyCustom           PrivacyValue = "CUSTOM"
	PrivacySelf             PrivacyValue = "SELF"
)

type Privacy struct {
	Value PrivacyValue `json:"value"`
	Allow string       `json:"allow"`
	Deny  string       `json:"deny"`
}

func (p *Privacy) JSON() string {
	d, _ := json.Marshal(p)
	return string(d)
}

type Option struct {
	Privacy *Privacy
}

// UploadSession facebook upload session struct
type UploadSession struct {
	// ID of a fb resource, possible value are user, page, event, group.
	ID int64

	// AccessToken the token has permission to upload video to the fb resource.
	AccessToken string

	// Endpoint url to upload video
	Endpoint string

	Client *http.Client

	// FilePath location of the file in disk
	FilePath        string
	fileName        string
	fileExt         string
	fileData        []byte
	fileSize        int
	fileChunkNumber int
	fileChunkFolder string
	uploadSessionID string
}

// NewUploadSession create a new fb upload session.
func NewUploadSession(filePath string, fbResourceID int64, accessToken string) *UploadSession {
	uploadSession := &UploadSession{
		ID:          fbResourceID,
		AccessToken: accessToken,
		Endpoint:    fmt.Sprintf("https://graph-video.facebook.com/v2.6/%d/videos", fbResourceID),
		Client:      &http.Client{},
		FilePath:    filePath,
	}

	file, err := os.Open(uploadSession.FilePath)
	if err != nil {
		return uploadSession
	}
	defer file.Close()

	buf := bytes.Buffer{}
	io.Copy(&buf, file)

	uploadSession.fileData = buf.Bytes()
	uploadSession.fileExt = path.Ext(uploadSession.FilePath)
	uploadSession.fileName = strings.TrimSuffix(path.Base(uploadSession.FilePath), uploadSession.fileExt)
	uploadSession.fileSize = buf.Len()

	return uploadSession
}

// Upload upload the file to fb server
func (uploadSession *UploadSession) Upload(option Option) error {
	session, err := uploadSession.initialize(option)
	if err != nil {
		return err
	}

	uploadSession.fileChunkFolder = os.TempDir() + "/" + session.UploadSessionID
	if err := os.MkdirAll(uploadSession.fileChunkFolder, 0777); err != nil {
		return err
	}

	if err := uploadSession.uploadChunk(session.UploadSessionID, session.ChunkOffset); err != nil {
		return err
	}

	return uploadSession.finish(session.UploadSessionID)
}

// Initialize Create a new upload session
func (uploadSession *UploadSession) initialize(option Option) (*SessionInfo, error) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("access_token", uploadSession.AccessToken)
	writer.WriteField("upload_phase", "start")
	writer.WriteField("file_size", fmt.Sprintf("%d", uploadSession.fileSize))
	if option.Privacy != nil {
		writer.WriteField("privacy", option.Privacy.JSON())
	}
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, uploadSession.Endpoint, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	res, err := uploadSession.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Can not create new upload session, got error %s", NewErrorFromBody(res.Body).Struct.Message)
	}

	var sessionInfo SessionInfo
	if err := json.NewDecoder(res.Body).Decode(&sessionInfo); err != nil {
		return nil, err
	}

	return &sessionInfo, nil
}

// Upload transfer chunk file to facebook server
func (uploadSession *UploadSession) uploadChunk(uploadSessionID string, chunkOffset *ChunkOffset) error {
	uploadSession.fileChunkNumber++

	contentType, chunkPart, err := uploadSession.createNewChunk(uploadSessionID, chunkOffset)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest(http.MethodPost, uploadSession.Endpoint, chunkPart)
	req.Header.Add("Content-Type", contentType)

	res, err := uploadSession.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Can not upload chunk %d, got error %s", uploadSession.fileChunkNumber, NewErrorFromBody(res.Body).Struct.Message)
	}

	var newChunkOffset ChunkOffset
	if err := json.NewDecoder(res.Body).Decode(&chunkOffset); err != nil {
		return err
	}

	if newChunkOffset.StartOffset != newChunkOffset.EndOffset {
		return uploadSession.uploadChunk(uploadSessionID, &newChunkOffset)
	}

	return nil
}

// Finish finish the upload session and post the uploaded video to fb resource.
func (uploadSession *UploadSession) finish(uploadSessionID string) error {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("access_token", uploadSession.AccessToken)
	writer.WriteField("upload_phase", "finish")
	writer.WriteField("upload_session_id", uploadSessionID)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, uploadSession.Endpoint, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := uploadSession.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Can not post uploaded video, got error %s", NewErrorFromBody(res.Body).Struct.Message)
	}

	return nil
}

func (uploadSession *UploadSession) createNewChunk(uploadSessionID string, chunkOffset *ChunkOffset) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	startOffset, _ := strconv.ParseInt(chunkOffset.StartOffset, 10, 64)
	endOffset, _ := strconv.ParseInt(chunkOffset.EndOffset, 10, 64)

	fileChunk := uploadSession.fileData[startOffset:endOffset]
	fileChunkName := fmt.Sprintf("@%s%d%s", uploadSession.fileName, uploadSession.fileChunkNumber, uploadSession.fileExt)
	fileChunkPath := uploadSession.fileChunkFolder + "/" + fileChunkName
	if _, err := os.Create(fileChunkPath); err != nil {
		return "", nil, err
	}

	part, err := writer.CreateFormFile("video_file_chunk", fileChunkPath)
	if err != nil {
		return "", nil, err
	}
	part.Write(fileChunk)

	writer.WriteField("access_token", uploadSession.AccessToken)
	writer.WriteField("upload_phase", "transfer")
	writer.WriteField("upload_session_id", uploadSessionID)
	writer.WriteField("start_offset", chunkOffset.StartOffset)

	if err = writer.Close(); err != nil {
		return "", nil, err
	}

	return writer.FormDataContentType(), body, nil
}
