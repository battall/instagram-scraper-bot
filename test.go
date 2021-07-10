package main

import (
	"log"

  "net/http"
  "encoding/json"

  "./instapi"
)

var HTTP_PORT = ":9000"

func main() {
  log.Println("Starting")

  session := instapi.Session{Id: "8469245135", Username: "smy.clsn", Password: "kemalbosinsan",
		CookiePath: "cookies/8469245135",
		WebUA:      "Mozilla/5.0 (Linux; Android 7.0; SM-G892A Build/NRD90M; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/67.0.3396.87 Mobile Safari/537.36",
		AppUA:      "Instagram 27.0.0.7 " + instapi.USERAGENTS[0]}

	if err := session.Login(); err != nil {
	  log.Fatal("Error Login:", err.Error())
	}

	log.Println("Logined, Fresh login:", session.FreshLogin)

  /*request := instapi.Fetch(session).SetResource("info", "253106136").Send().ToText()
  if request.Error != nil {
  	log.Fatal("Error Request: ", request.Error, request.Response.Header)
  }
  
  log.Println(request.Text)*/
//
  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")

    query := r.URL.Query()

    var result map[string]interface{};
    switch method := query.Get("method"); method {
      case "info":
        request := instapi.Fetch(session).SetResource("info", query.Get("id")).Send().ToJSON()
        if request.Error != nil {
  	      http.Error(w, request.Error.Error(), request.ErrorCode) 
          return
        }
        result = request.JSON;
      case "followers":
        request := instapi.Fetch(session).SetResource("followers_web", query.Get("id"), query.Get("max_id")).Send().ToJSON()
        if request.Error != nil {
  	      http.Error(w, request.Error.Error(), request.ErrorCode) 
          return
        }
        result = request.JSON;
    }
    
    js, err := json.Marshal(result)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError) 
      return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
  })
  http.ListenAndServe(HTTP_PORT, nil)
}
