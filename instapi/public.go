package instapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)


/*func GetUsernamePublic(id uint64) (string, error){

}*/


// Get new token from instagram, when logining you need this.
func GetCSRFToken() (string, error) { //new token
	response, err := http.Get("https://i.instagram.com/api/v1/users/")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", errors.New(response.Status)
	}
	for _, cookie := range response.Cookies() {
    if (cookie.Name == "csrftoken") {
    	return cookie.Value, nil
    }
  }
	return "", errors.New("CSRFToken not sent")
}

// Get Public info of any account
func GetInfo(id uint64, proxy string, timeout time.Duration) (User, error) {
	info := User{Id: id}
	//netClient, err := CreateClient(proxy, timeout)
	//if err != nil {
	//	return info, err
	//}

	response, err := http.DefaultClient.Get("https://i.instagram.com/api/v1/users/" + strconv.FormatInt(int64(id), 10) + "/info/")
	if err != nil {
		return info, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200:
		var result map[string]interface{}
		json.NewDecoder(response.Body).Decode(&result)
		tempInfo := result["user"].(map[string]interface{})
		info.Username = tempInfo["username"].(string)

		//second request start
		response2, err := http.DefaultClient.Get("https://www.instagram.com/" + info.Username + "/?__a=1")
		if err != nil {
			return info, err
		}
		defer response2.Body.Close()

		switch response2.StatusCode {
		case 200:
			json.NewDecoder(response2.Body).Decode(&result)
			tempInfo2 := result["graphql"].(map[string]interface{})["user"].(map[string]interface{})
			info.Full_name = tempInfo2["full_name"].(string)
			info.Is_private = tempInfo2["is_private"].(bool)
			info.Biography = tempInfo2["biography"].(string)
			if tempInfo2["external_url"] != nil {
				info.External_url = tempInfo2["external_url"].(string)
			}
			return info, nil
		default:
			return info, errors.New("2-" + strconv.FormatInt(int64(response2.StatusCode), 10))
		}
	case 404:
		info.Is_disabled = true
		return info, nil
	default:
		return info, errors.New("1-" + strconv.FormatInt(int64(response.StatusCode), 10))
	}

}
