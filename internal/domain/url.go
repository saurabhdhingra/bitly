package domain

import (
	"time"
)

type URL struct{
	ID			string		`json:"id" bson:"_id.omitempty"`
	URL			string		`json:"url" bson:"url"`
	ShortCode	string		`json:"shortcode" bson:"shortCode"`
	CreatedAt	time.Time	`json:"createdAt" bson:"createdAt"`
	UpdatedAt	time.Time	`json:"updatedAt" bson:"updatedAt"`
	AccessCount	int64		`json:"accessCount" bson:"accessCount"`
}
