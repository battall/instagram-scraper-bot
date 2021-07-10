package instapi

type User struct {
	Id             uint64
	Username       string
	Full_name      string
	Is_private     bool
	Biography      string
	External_url   string
	Is_disabled    bool
	Profile_pic_id string
}

type Media struct {
	Id         string
	User_id    string
	Taken_at   int
	Media_type int
	Feed_type  int
	Urls       []string
	Carousel_media_types []int
}