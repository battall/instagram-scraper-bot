package main

import (
	"./instapi"
	
  "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"

	"net/http"
	"io"
	"os"
  "strconv"
)

var mediaExt = [3]string{"", ".jpg", ".mp4"}
var methodId = map[string]string{"info": "1", "feed": "2", "reel": "3"}

// Want to do a struct because of the provided flexibility 
// And cleaner global scope for the future
type BotManager struct {
	// Database variables
  MongoURI string
  MongoURIDB string
  CTX      context.Context
  Database *mongo.Database
  
  // Bot Variables
	Sessions map[string]instapi.Session
	Followings map[string][]string
}

func (manager *BotManager) Init() {
  manager.CTX = context.Background()//context.WithTimeout(context.Background(), 10*time.Second)
  
  client, _ := mongo.Connect(manager.CTX, options.Client().ApplyURI(manager.MongoURI))
  manager.Database = client.Database(manager.MongoURIDB)

  botManager.Followings = make(map[string][]string)
  botManager.Sessions = make(map[string]instapi.Session)
}

// Gets all accounts from mongodb
func (manager *BotManager) GetAccounts() ([]map[string]interface{}, error){
  collection := manager.Database.Collection("users")

	// Find() method returns cursor, we are gonna itarete over it
	cur, err := collection.Find(manager.CTX, bson.M{"password": bson.M{"$exists": true}})
	if err != nil {
		return nil, err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	var results []map[string]interface{}
	for cur.Next(manager.CTX) {
		// create a value into which the single document can be decoded
		var elem map[string]interface{}
		if err := cur.Decode(&elem); err != nil {
			return nil, err
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	cur.Close(manager.CTX)

	return results, nil
}

// CHECK OPERATIONS
func (manager *BotManager) GetUserForCheck(method string) (string, error) {
	collection := manager.Database.Collection("users")

	options := options.FindOne()
	options.SetSort(bson.D{{"last_checked." + methodId[method], 1}})
	filter := bson.M{"last_checked.0": 1}
  // if method is feed or reel, user account must be active.
	if method == "feed" || method == "reel" {
		filter["is_disabled"] = false
	}

	var user map[string]interface{}
	err := collection.FindOne(manager.CTX, filter, options).Decode(&user)
	if err != nil {
		return "", err
	}
	return user["_id"].(string), nil
}

func (manager *BotManager) UpdateCheckedUser(method string, userId string) error {
	collection := manager.Database.Collection("users")

	_, err := collection.UpdateOne(manager.CTX, bson.M{"_id": userId}, bson.M{
		"$currentDate": bson.M{
			"last_checked." + methodId[method]: true,
		},
	})

	return err
}

func (manager *BotManager) GetAllCheckings() ([]string, error) {
  collection := manager.Database.Collection("users")

  // Find() method returns cursor, we are gonna itarete over it
  cur, err := collection.Find(manager.CTX, bson.M{"password": bson.M{"$exists": false}})
  if err != nil {
    return nil, err
  }

  // Finding multiple documents returns a cursor
  // Iterating through the cursor allows us to decode documents one at a time
  var results []map[string]interface{}
  for cur.Next(manager.CTX) {
    // create a value into which the single document can be decoded
    var elem map[string]interface{}
    if err := cur.Decode(&elem); err != nil {
      return nil, err
    }
    results = append(results, elem)
  }

  if err := cur.Err(); err != nil {
    return nil, err
  }

  cur.Close(manager.CTX)
  
  results_id := make([]string, len(results));
  for i, user := range results {
    results_id[i] = user["_id"].(string)
  }

  return results_id, nil
}

// END MONGO OPERATIONS
// AFTER HERE STARTS SOME BITCHING

func (manager *BotManager) GetSessionForCheck(userId string) string {
	for sessionId, userList := range manager.Followings {
		for _, u := range userList {
			if userId == u {
				return sessionId
			}
		}
	}
	return ""
}


func (manager *BotManager) AddSession(session instapi.Session, followings []string) error {
  // Add followings
  manager.Followings[session.Id] = followings;
  
  // Add session
  manager.Sessions[session.Id] = session

  return nil
}

// Before saving checks for exists in mongo,
// if not exists, downloads, after it creates
// media in mongo
func (manager *BotManager) SaveMedia(media instapi.Media) error {
  collection := manager.Database.Collection("media")

  // Check document exists
  count, err := collection.CountDocuments(manager.CTX, bson.M{"_id": media.Id})
  if err != nil || count > 0 {// If count more than 0, it will return "nil".
  	return err
  }
  
  // Download and save document
  for i, url := range media.Urls {
  	res, err := http.Get(url)
    if err != nil {
    	return err
    }
    defer res.Body.Close()
    
    name := "/home/battal/Desktop/ig_medias/" + media.Id
    if media.Media_type == 8 {
    	name += "_" + strconv.Itoa(i) + mediaExt[media.Carousel_media_types[i]]
    } else {
    	name += mediaExt[media.Media_type]
    }

    out, err := os.Create(name)
    if err != nil {
      return err
    }
    defer out.Close()

    io.Copy(out, res.Body)
  }

  // Add document to MongoDB
  _, err = collection.InsertOne(manager.CTX, bson.D{
    {"_id", media.Id},
    {"user_id", media.User_id},
    {"taken_at", media.Taken_at},
    {"media_type", media.Media_type},
    {"feed_type", media.Feed_type},
  })
  
  if err != nil {
  	return err;
  }

  return nil
}

/*
return this.database.models.media.countDocuments({
      _id: media._id
    })
    .then(count => {
      if (count === 1) return "alredyInDb";
      console.log(media);
      if (media.media_type == 8) {
        var query = [];
        for (let i = 0, len = media.url.length; i < len; i++) query.push(this.media.putURL('ig/' + media._id + '_' + i + mediaExt[media.carousel_media_types[i]], media.url[i]));
        return Promise.all(query)
      } else {
        return this.media.putURL('ig/' + media._id + mediaExt[media.media_type], media.url)
      }
    })
    .then(msg => msg !== "alredyInDb" ? this.database.models.media.create(media) : "")
    .catch(error => Promise.reject(error))*/