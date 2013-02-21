package api

// This file contains the data schema returned by the Wobble API Endpoint. Designed to be used with `json.Unmarshal()`.

import "fmt"

// The generic API error
type WobbleApiError struct {
	Code    int
	Message string
}

func (err WobbleApiError) Error() string {
	return fmt.Sprintf("%d: %s", err.Code, err.Message)
}

type SearchResultTopic struct {
	TopicId         string `json:"id"`
	Abstract        string // First two lines of the root post (id: 1)
	PostCountUnread int    `json:"post_count_unread"` // Number of unread posts
	PostCountTotal  int    `json:"post_count_total"`

	Archived int `json:"archived"`
}

type Topic struct {
	TopicId  string         `json:"id"`
	Messages []TopicMessage `json:"messages"`

	Readers []User `json:"readers"` // List of all users that are currently allowed to see this topic
	Writers []User `json:"writers"` // List of all users that have posts in this topic (might not be part of the topic anymore)

	Posts     []Post  `json:"posts"`
	Archived  int     `json:"archived"`   // Is this topic archived? 1 = true, 0 = false
	CreatedAt float64 `json:"created_at"` //

}

type TopicMessage struct {
	MessageId string      `json:"message_id"`
	Message   interface{} `json:"message"`
}

type Post struct {
	PostId       string  `json:"id"`
	Content      *string `json:"content"`
	ParentPostId *string `json:"parent"`

	Timestamp    float64 // Unixtimestamp
	CreatedAt    float64 `json:"created_at"`
	RevisionNo   int     `json:"revision_no"`
	Deleted      int     `json:"deleted"`
	Unread       int     `json:"unread"`
	IntendedPost int     `json:"intended_post"`
	Lock         *Lock   `json:"locked"`
	Users        []int   `json:"users"`
}

type Lock struct {
	UserId int `json:"user_id"`
}

type User struct {
	UserId          int    `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Online          int    `json:"online"`
	GravatarImgHash string `json:"img"`
}

type UserSignoutResult struct{}

type UserLoginResult struct {
	ApiKey string `json:"apikey"`
}

type ListTopicResponse struct {
	InboxUnreadTopics int                 "inbox_unread_topics"
	Topics            []SearchResultTopic "Topics"
}

type EditPostResponse struct {
	RevisionNo float64 `json:"revision_no"`
}

type Notification struct {
	Type    string `json:"type"`
	UserId  int    `json:"user_id"`
	TopicId string `json:"topic_id"`
	PostId  string `json:"post_id"`
}

type GetNotificationsResponse struct {
	Messages      []Notification `json:"messags"`
	NextTimestamp float64        `json:"next_timestamp"`
}
