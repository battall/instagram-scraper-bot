package instapi

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

func Fetch(session ...Session) *RequestBuilder {
	rb := &RequestBuilder{}
	if len(session) > 0 {
		rb.Session = session[0]
	}
	return rb
}

type RequestBuilder struct {
	Method string
	URL    string

	Session Session

	Request  *http.Request
	Response *http.Response
  
  Error error
  ErrorCode int
	
	Text string
	JSON map[string]interface{}
}

// Sets resource from route
// puts params to {} values on url
// checks for session and headers
func (rb *RequestBuilder) SetResource(name string, params ...string) *RequestBuilder {
	// Prepare URL details
	route := Routes[name]
	rb.URL = route.URL
	rb.Method = route.Method
	for _, element := range params {
		rb.URL = strings.Replace(rb.URL, "{}", element, 1)
	}
    // Delete other {}
  rb.URL = strings.Replace(rb.URL, "{}", "", -1)
    
  rb.createRequest()
	if rb.Error != nil {
		return rb
	}

	// Check for session and headers
	if route.RequireSession && !rb.Session.Authenticated {
		rb.Error = errors.New("this endpoint needs authenticated session")
		return rb
	}

	if route.RequireAppHeaders {
		rb.setMobileHeaders()
	}

	return rb
}

func (rb *RequestBuilder) createRequest() *RequestBuilder {
	request, err := http.NewRequest(rb.Method, rb.URL, nil)
	if err != nil {
		rb.Error = err
		return rb
	}
	rb.Request = request
	return rb
}

func (rb *RequestBuilder) setMobileHeaders() *RequestBuilder {
	//rb.Request.Header.Set("Host", "i.instagram.com")
	rb.Request.Header.Set("Accept", "*/*")
	rb.Request.Header.Set("Accept-Language", "en-US")
	rb.Request.Header.Set("User-Agent", rb.Session.AppUA)
	rb.Request.Header.Set("X-Ig-Capabilities", "3QI=")
	rb.Request.Header.Set("X-Ig-Connection-Type", "WIFI")
	return rb
}

func (rb *RequestBuilder) Send() *RequestBuilder {
	if rb.Error != nil {
		return rb
	}

	// Do the request
	response, err := rb.Session.Client.Do(rb.Request)
	if err != nil {
		rb.Error = err
	} else if response.StatusCode != 200 {
		rb.ErrorCode = response.StatusCode
		rb.Error = errors.New("Unknown error")
		// If error response is json, put body to error too.
	  if value := response.Header.Get("content-type"); value == "application/json; charset=utf-8" {
			body, _ := ioutil.ReadAll(response.Body)
			rb.Error = errors.New(string(body))
		}
	}
	rb.Response = response

	return rb
}

func (rb *RequestBuilder) ToText() *RequestBuilder {
	if rb.Error != nil {
		return rb
	}

	body, err := ioutil.ReadAll(rb.Response.Body)
	if err != nil {
		rb.Error = err
		return rb
	}

	rb.Text = string(body)

	return rb
}

func (rb *RequestBuilder) ToJSON() *RequestBuilder {
	if rb.Error != nil {
		return rb
	}

	var result map[string]interface{}
	json.NewDecoder(rb.Response.Body).Decode(&result)

	rb.JSON = result

	return rb
}
