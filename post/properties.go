package post

type PostLocation struct {
	Latitude  string `json:"latitude" validate:"required"`
	Longitude string `json:"longitude" validate:"required"`
	Altitude  string `json:"altitude,omitempty"`
}

type PostContent struct {
	Html string `json:"html" validate:"required"`
	Text string `json:"text" validate:"required"`
}

type PostPhoto struct {
	Value string `json:"value" validate:"required,url"`
	Alt   string `json:"alt"`
}

// Checkin is an incomplete h-card which is used in checkin posts.
type Checkin struct {
	Type      string `json:"type" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Latitude  string `json:"latitude" validate:"required"`
	Longitude string `json:"longitude" validate:"required"`
	Altitude  string `json:"altitude,omitempty"`
}
