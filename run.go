package main

import (
	"log"
  "encoding/json"
  "net/http"
  "strconv"
  "strings"
  "time"
  "bufio"
  "os"

  "go.mongodb.org/mongo-driver/bson"

  "./instapi"
)

var HTTP_PORT = ":4000"

var botManager = BotManager{
	MongoURI: "mongodb:///ig?retryWrites=false",
  MongoURIDB: "ig",
}

func main() {
  http.HandleFunc("/accounts", func (w http.ResponseWriter, r *http.Request) {
    accounts, err := botManager.GetAccounts()
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError) 
      return
    }
    
    js, err := json.Marshal(accounts)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError) 
      return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
  })
  go http.ListenAndServe(HTTP_PORT, nil)
  
  log.Println("Started HTTP Server")
	
	// Done HTTP server Operations

  // Reader for input from user to add or not add user to checkings.
  reader := bufio.NewReader(os.Stdin)

  botManager.Init() // Connect to mongo etc

  accounts, err := botManager.GetAccounts()
  checkings, err := botManager.GetAllCheckings()
  if err != nil {
		log.Fatal(err)
	}
  
  log.Println("Got", len(accounts), "accounts")
  
  // Login accounts and get following list.
	for _, account := range accounts {
		session := instapi.Session{Id: account["_id"].(string), Username: account["username"].(string), Password: account["password"].(string),
			CookiePath: "cookies/" + account["_id"].(string),
			WebUA:      "Mozilla/5.0 (Linux; Android 7.0; SM-G892A Build/NRD90M; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/67.0.3396.87 Mobile Safari/537.36",
			AppUA:      "Instagram 27.0.0.7 " + instapi.USERAGENTS[0]}

		if err = session.Login(); err != nil {
			log.Println("Error logining:", err.Error() + ".",
			  session.Id, session.Username, session.Password)
      continue
		}

    // Load followings
		request := instapi.Fetch(session).SetResource("following", session.Id).Send().ToJSON()
		if request.Error != nil {
			log.Println("Error loading followings:", request.Error.Error() + ".",
			  session.Id, session.Username, session.Password)
      continue
		}
		
		// Parse response, push to array
		followings := make([]string, len(request.JSON["users"].([]interface{})))
		for i, user := range request.JSON["users"].([]interface{}) {
			user := user.(map[string]interface{})
			id := strconv.FormatFloat(user["pk"].(float64), 'f', -1, 64)
			followings[i] = id

      // CHECK FOR IF FOLLOWINGS IN CHECKINGS
		  found := false;
		  for _, checking_id := range checkings {
        if checking_id == id {
        	found = true
        }
		  }
		  if !found {
		  	req, _ := http.NewRequest("GET", "https://i.instagram.com/api/v1/users/" + id + "/info/", nil)
				req.Header.Set("User-Agent", "Instagram 19.0.0.27.91 (iPhone6,1; iPhone OS 9_3_1; en_US; en; scale=2.00; gamut=normal; 640x1136) AppleWebKit/420+")
				res, _ := http.DefaultClient.Do(req)
	      defer res.Body.Close()

	      var info map[string]interface{}
        json.NewDecoder(res.Body).Decode(&info)

		  	log.Println("Not in list", id, info["user"].(map[string]interface{})["username"].(string), "Add?")
		  	text, _ := reader.ReadString('\n')
		  	text = strings.Replace(text, "\n", "", -1)
		  	user2add := bson.M{"_id": id, "is_disabled": false};
		  	if text == "y" {
		  		user2add["last_checked"] = [4]int{1, 0, 0, 0}
		  		log.Println("Adding")
		  	} else if text == "n" {
		  		user2add["last_checked"] = [4]int{0, 0, 0, 0}
		  		log.Println("Blacklist")
		  	} else {
		  		panic("Wrong input")
		  	}
		  	_, err = botManager.Database.Collection("users").InsertOne(botManager.CTX, user2add)
		  	if err != nil {
		  		log.Println("MongoError", err)
		  	}
		  }
		  // END CHECK FOR NOT IN CHECKINGS
		}

		log.Println("Logined", session.Id, session.Username, session.Password,
		  "-", "FreshLogin:", session.FreshLogin, "Following count:", len(followings))

		botManager.AddSession(session, followings)
	}
  
  // Checking starts here

  for {
		user, err := botManager.GetUserForCheck("reel")
		if err != nil {
			log.Println("MongoError GetUserForCheck", err)
			continue
		}

		sessionId := botManager.GetSessionForCheck(user)
		if sessionId == "" {
			log.Println(user, "No one is following")
			if err := botManager.UpdateCheckedUser("reel", user); err != nil {
				log.Println("MongoError", err)
			}
			continue
		}

		session := botManager.Sessions[sessionId]

		request := instapi.Fetch(session).SetResource("story", user).Send().ToJSON()
		if request.Error != nil && request.ErrorCode == 429{
			log.Println(user, "Request error", request.ErrorCode, request.Error)
			time.Sleep(30 * time.Second)
			continue
		} else if request.Error != nil {
			log.Println(user, "Request error", request.ErrorCode, request.Error)
		} else if request.JSON["reel"] == nil {
		  log.Println(user, "Media count 0")
		} else {
			items := request.JSON["reel"].(map[string]interface{})["items"].([]interface{})
			medias := make([]instapi.Media, len(items))

			for i, item := range items {
        item := item.(map[string]interface{})

        id_index := strings.Index(item["id"].(string), "_")
        url_base := item[[2]string{"image_versions2", "video_versions"}[int(item["media_type"].(float64)) - 1]]
        if int(item["media_type"].(float64)) == 1 {
        	url_base = url_base.(map[string]interface{})["candidates"]
        }

        medias[i] = instapi.Media{
          Id: item["id"].(string)[0:id_index],
          User_id: user,
          Taken_at: int(item["taken_at"].(float64)),
          Media_type: int(item["media_type"].(float64)),
          Feed_type: 1,
          Urls: []string{url_base.([]interface{})[0].(map[string]interface{})["url"].(string)},
        }

        err = botManager.SaveMedia(medias[i])
        if err != nil {
        	log.Println("Error", err)
        	continue
        }

        //log.Println(medias[i])
			}

			log.Println(user, "Media count", len(items))
		}
		
		if err := botManager.UpdateCheckedUser("reel", user); err != nil {
			log.Println("MongoError", err)
		}
	}
}

