package instapi

import (
	"errors"
	"os"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
  
  "net/url"
	"net/http"
	"net/http/cookiejar"
  "golang.org/x/net/http2"
)

type Session struct {
	Id       string //uint64
	Username string
	Password string
  
  Timeout time.Duration
  ProxyURL string

	Authenticated bool
	FreshLogin    bool
	CSRFToken     string
	CookiePath    string
	Client        *http.Client

	WebUA string
	AppUA string
}

func (session *Session) Login() error {

	//client.

	// Create http client
	session.Client = &http.Client{}
	session.Client.Timeout = session.Timeout
	session.Client.Jar, _ = cookiejar.New(nil)
	session.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
  
  session.Client.Transport = &http2.Transport{}

  // Check proxy specified
  if session.ProxyURL != "" {
    proxy, err := url.Parse(session.ProxyURL)
    if err != nil {
    	return err
    }
  	session.Client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxy) // set proxy
    //session.Client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl
  }
  
	// If dont have cookie path, fresh login
	if session.CookiePath == "" {
		return session.freshLogin()
	}

	// If file not exists, fresh login
	if _, err := os.Stat(session.CookiePath); os.IsNotExist(err) {
		return session.freshLogin()
	}

	// Try to read cookies
	data, err := ioutil.ReadFile(session.CookiePath)
	if err != nil {
		return err
	}

	// Try to load cookies
	cookies := new([]*http.Cookie)
	if err := json.Unmarshal(data, cookies); err != nil {
		return err
	}

	// On here TLD_URL must be www.instagram.com because edit doesnt work otherwise
	// After requesting "edit", cookies appending i.instagram.com too
	// And works fine
	session.Client.Jar.SetCookies(WEB_HOST_URL, *cookies)
	session.Client.Jar.SetCookies(HOST_URL, *cookies)

	// If cant get edit (profile info), fresh login
	request := Fetch(*session).SetResource("edit").Send().ToJSON()
	if request.Error != nil  {
		return session.freshLogin()
	}
  
	// Check username confuse
	form_data := request.JSON["form_data"].(map[string]interface{})
	if form_data["username"] != session.Username {
		return errors.New("username mismatch, got '" + session.Username + "' but account has '" + form_data["username"].(string) + "'")
	}

	session.Authenticated = true

	session.SaveCookies()
	return nil
}

func (session *Session) freshLogin() error {
	session.FreshLogin = true

	// Clear cookies
	session.Client.Jar, _ = cookiejar.New(nil)

	// Get new CSRF token
	CSRFToken, err := GetCSRFToken()
	if err != nil {
		return err
	}
	session.CSRFToken = CSRFToken

	// Login request
	request, _ := http.NewRequest("POST", "https://www.instagram.com/accounts/login/ajax/",
		strings.NewReader(`username=`+session.Username+`&password=`+session.Password))
	request.Header.Set("X-CSRFToken", session.CSRFToken)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("User-Agent", session.WebUA)

	response, err := session.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Parse response
	var result map[string]interface{}
	json.NewDecoder(response.Body).Decode(&result)

	if result["status"] != "ok" {
		return errors.New(response.Status)
	}

	if !result["user"].(bool) {
		return errors.New("404")
	}

	if !result["authenticated"].(bool) {
		return errors.New("401")
	}

	session.Authenticated = result["authenticated"].(bool)
	session.Id = result["userId"].(string)

	session.SaveCookies()
	return nil
}

func (session *Session) SaveCookies() error {
	data, err := json.MarshalIndent(session.Client.Jar.Cookies(WEB_HOST_URL), "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(session.CookiePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
