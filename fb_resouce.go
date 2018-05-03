package fbvideo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

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

	if res.StatusCode > 299 {
		var fbErr Error
		json.NewDecoder(res.Body).Decode(&fbErr)
		return nil, errors.New(fbErr.Struct.Message)
	}

	var m Map
	json.NewDecoder(res.Body).Decode(&m)

	return m, nil
}

func GetResourceInfo(resourseID string, accessToken string) (Map, error) {
	endpoint := fmt.Sprintf("https://graph.facebook.com/%s?access_token=%s", resourseID, accessToken)

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 299 {
		var fbErr Error
		json.NewDecoder(res.Body).Decode(&fbErr)
		return nil, errors.New(fbErr.Struct.Message)
	}

	var m Map
	json.NewDecoder(res.Body).Decode(&m)

	return m, nil
}
