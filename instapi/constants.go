package instapi

import (
	"net/url"
)

const TLD = "instagram.com"
const HOSTNAME = "i." + TLD
const WEB_HOSTNAME = "www." + TLD
const HOST = "https://" + HOSTNAME + "/"
const WEB_HOST = "https://" + WEB_HOSTNAME + "/"

var WEB_HOST_URL, _ = url.Parse(WEB_HOST)
var HOST_URL, _ = url.Parse(HOST)

var USERAGENTS = [2]string{
	"Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G920F; zeroflte; samsungexynos7420)",
	"(iPhone6,1; iPhone OS 9_3_1; en_US; en; scale=1.00; gamut=normal; 1440x2560) AppleWebKit/420+",
}

type Route struct {
	URL               string
	Method            string

	RequireBody       bool
	RequireSession    bool
	RequireAppHeaders bool
	RequireSignedData bool
}

// Rule: if its len ==
var Routes = map[string]Route{
	// Web endpoints
	"edit": Route{URL: WEB_HOST + "accounts/edit/?__a=1"},

	// Needs signed signature
	"reels": Route{URL: HOST + "api/v1/feed/reels_media/", Method: "POST", RequireSession: true, RequireAppHeaders: true, RequireSignedData: true},

	// These are mobile endpoints
	"info":       Route{URL: HOST + "api/v1/users/{}/info/", RequireAppHeaders: true},
	"feed":       Route{URL: HOST + "api/v1/feed/user/{}/?max_id={}", RequireSession: true, RequireAppHeaders: true},
	"story":      Route{URL: HOST + "api/v1/feed/user/{}/story/", RequireSession: true, RequireAppHeaders: true},
	"highlights": Route{URL: HOST + "api/v1/highlights/{}/highlights_tray/", RequireSession: true, RequireAppHeaders: true},

	"media": Route{URL: HOST + "api/v1/media/{}/info/", RequireSession: true, RequireAppHeaders: true},
  

  "followers_web": Route{URL: WEB_HOST + `graphql/query/?query_hash=c76146de99bb02f6415203be841dd25a&variables={"id":"{}","include_reel":false,"fetch_mutual":false,"first":50,"after":"{}"}`, RequireSession: true},
  
	"followers": Route{URL: HOST + "api/v1/friendships/{}/followers/?max_id={}", RequireSession: true, RequireAppHeaders: true},
	"following": Route{URL: HOST + "api/v1/friendships/{}/following/", RequireSession: true, RequireAppHeaders: true},
}
