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
	ID string

	// AccessToken the token has permission to upload video to the fb resource.
	AccessToken string

	// Endpoint url to upload video
	Endpoint  string
	Transport http.RoundTripper

	// FilePath location of the file in disk
	FilePath        string
	fileName        string
	fileExt         string
	file            *os.File
	fileSize        int64
	fileChunkFolder string
	fileChunkNumber uint
	uploadSessionID string
}

// NewUploadSession create a new fb upload session.
func NewUploadSession(filePath string, fbResourceID string, accessToken string) *UploadSession {
	uploadSession := &UploadSession{
		ID:          fbResourceID,
		AccessToken: accessToken,
		Endpoint:    fmt.Sprintf("https://graph-video.facebook.com/v2.6/%s/videos", fbResourceID),
		Transport:   http.DefaultTransport,
		FilePath:    filePath,
	}

	file, err := os.Open(uploadSession.FilePath)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	uploadSession.file = file
	uploadSession.fileExt = path.Ext(uploadSession.FilePath)
	uploadSession.fileName = strings.TrimSuffix(path.Base(uploadSession.FilePath), uploadSession.fileExt)
	uploadSession.fileSize = fileInfo.Size()

	return uploadSession
}

// Upload upload the file to fb server
func (uploadSession *UploadSession) Upload(option Option) (string, error) {
	defer uploadSession.file.Close()

	session, err := uploadSession.initialize(option)
	if err != nil {
		return "", err
	}

	uploadSession.fileChunkFolder = "testdata" + "/fb" + session.UploadSessionID
	if err := os.MkdirAll(uploadSession.fileChunkFolder, 0777); err != nil {
		return "", err
	}
	defer os.RemoveAll(uploadSession.fileChunkFolder)

	if err := uploadSession.uploadChunk(session.UploadSessionID, session.ChunkOffset); err != nil {
		return "", err
	}

	if err := uploadSession.finish(session.UploadSessionID); err != nil {
		return "", err
	}

	return session.VideoID, nil
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

	res, err := uploadSession.Transport.RoundTrip(req)
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

// uploadChunk transfer chunk file to facebook server
func (uploadSession *UploadSession) uploadChunk(uploadSessionID string, chunkOffset *ChunkOffset) error {

	chunkInfo, err := uploadSession.createNewChunk(uploadSessionID, chunkOffset)
	if err != nil {
		return err
	}
	defer os.Remove(chunkInfo.Path)

	req, _ := http.NewRequest(http.MethodPost, uploadSession.Endpoint, chunkInfo.Body)
	req.Header.Add("Content-Type", chunkInfo.ContentType)
	req.Header.Add("Content-Length", fmt.Sprintf("%d", uploadSession.fileSize))

	res, err := uploadSession.Transport.RoundTrip(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Can not upload chunk file, got error %s", NewErrorFromBody(res.Body).Struct.Message)
	}

	var newChunkOffset ChunkOffset
	if err := json.NewDecoder(res.Body).Decode(&newChunkOffset); err != nil {
		return err
	}

	if newChunkOffset.StartOffset != newChunkOffset.EndOffset {
		os.Remove(chunkInfo.Path)
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
	res, err := uploadSession.Transport.RoundTrip(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Can not post uploaded video, got error %s", NewErrorFromBody(res.Body).Struct.Message)
	}

	return nil
}

func (session *UploadSession) createNewChunk(uploadSessionID string, chunkOffset *ChunkOffset) (*ChunkInfo, error) {
	session.fileChunkNumber++

	startOffset, _ := strconv.ParseInt(chunkOffset.StartOffset, 10, 64)
	endOffset, _ := strconv.ParseInt(chunkOffset.EndOffset, 10, 64)
	if endOffset > session.fileSize {
		endOffset = session.fileSize
	}

	fileChunkName := fmt.Sprintf("@chunk%d%s", session.fileChunkNumber, session.fileExt)
	fileChunkPath := session.fileChunkFolder + "/" + fileChunkName

	fileChunkBytes := make([]byte, endOffset-startOffset)
	if _, err := session.file.Read(fileChunkBytes); err != nil && err != io.EOF {
		return nil, err
	}

	fileChunk, err := os.Create(fileChunkPath)
	if err != nil {
		return nil, err
	}
	if _, err := fileChunk.Write(fileChunkBytes); err != nil {
		fileChunk.Close()
		return nil, err
	}
	fileChunk.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("video_file_chunk", fileChunkName)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(fileChunkBytes); err != nil {
		return nil, err
	}

	writer.WriteField("access_token", session.AccessToken)
	writer.WriteField("upload_session_id", uploadSessionID)
	writer.WriteField("upload_phase", "transfer")
	writer.WriteField("start_offset", chunkOffset.StartOffset)

	if err = writer.Close(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	chunkInfo := &ChunkInfo{
		Path:        fileChunkPath,
		Body:        body,
		ContentType: writer.FormDataContentType(),
	}

	return chunkInfo, nil
}
