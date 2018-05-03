package main

import (
	"fmt"
	"log"

	"github.com/canhlinh/go-fbvideo"
)

func main() {
	// You have to export your acess token to FB_ACCESS_TOKEN variable
	// run command: export FB_ACCESS_TOKEN="Your access token"
	accessToken := fbvideo.GetAccessToken()
	me, err := fbvideo.GetMe(accessToken)
	if err != nil {
		panic(err)
	}
	meID := me["id"].(string)

	uploadSession := fbvideo.NewUploadSession("../testdata/SampleVideo_720x480_10mb.mp4", meID, accessToken)

	// Uploads the video to your wall, only you can see it
	option := fbvideo.Option{Privacy: &fbvideo.Privacy{Value: fbvideo.PrivacySelf}}
	videoID, err := uploadSession.Upload(option)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("video_id %s", videoID)
}
