package fbvideo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Map map[string]interface{}

func GetAccessToken() string {
	return os.Getenv("FB_ACCESS_TOKEN")
}

func GetMe(accessToken string) (Map, error) {
	endpoint := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", accessToken)

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var m Map
	json.NewDecoder(res.Body).Decode(&m)
	return m, nil
}

func GetResourceInfo(resourseID int64, accessToken string) (Map, error) {
	endpoint := fmt.Sprintf("https://graph.facebook.com/%d?access_token=%s", resourseID, accessToken)

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var m Map
	json.NewDecoder(res.Body).Decode(&m)
	return m, nil
}
