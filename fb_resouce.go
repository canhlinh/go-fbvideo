package fbvideo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Map map[string]interface{}

func GetAccessToken() string {
	accessToken := os.Getenv("FB_ACCESS_TOKEN")
	if accessToken == "" {
		panic("access_token can not be empty")
	}

	return accessToken
}

func GetMe(accessToken string) (Map, error) {
	endpoint := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", accessToken)

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var m Map
	json.NewDecoder(res.Body).Decode(&m)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%v", m["error"])
	}

	return m, nil
}

func GetResourceInfo(resourseID string, accessToken string) (Map, error) {
	endpoint := fmt.Sprintf("https://graph.facebook.com/%s?access_token=%s", resourseID, accessToken)

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var m Map
	json.NewDecoder(res.Body).Decode(&m)
	return m, nil
}
