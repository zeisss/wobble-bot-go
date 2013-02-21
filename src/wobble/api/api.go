package api

import "log"
import "encoding/json"

func NewClient(endpoint string) (c *Client) {
	c = new(Client)
	c.endpoint = endpoint
	c.idSeq = 1
	return
}

type Client struct {
	idSeq     int
	endpoint  string
	sessionId string
}

func (this *Client) WobbleVersion() (string, error) {
	var params map[string]interface{} = make(map[string]interface{})
	var result string
	err := this.callMap("wobble.api_version", params, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

// Performs a request to the server to obtain an ApiKey.
func (this *Client) Login(email, password string) error {
	var params map[string]interface{} = make(map[string]interface{})
	params["email"] = email
	params["password"] = password

	var result UserLoginResult
	err := this.callMap("user_login", params, &result)
	if err != nil {
		return err
	}

	this.sessionId = result.ApiKey
	return nil
}

// Ends the current session
func (this *Client) Logout() error {
	var params map[string]interface{} = make(map[string]interface{})
	var result UserSignoutResult
	err := this.callMap("user_signout", params, &result)
	if err != nil {
		return err
	}
	this.sessionId = ""
	return nil
}

// Gets the current user
func (this *Client) GetCurrentUser() (*User, error) {
	var params map[string]interface{} = make(map[string]interface{})
	var result User
	err := this.callMap("user_get", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *Client) GetCurrentUserId() (int, error) {
	var params map[string]interface{} = make(map[string]interface{})
	var result int
	err := this.callMap("user_get_id", params, &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (this *Client) SubscribeNotifications() *Subscription {
	sub := newSubscription()
	go sub.loop(this)
	return sub
}

func (this *Client) getNotifications(next_timestamp float64) (*GetNotificationsResponse, error) {
	var params map[string]interface{} = make(map[string]interface{})
	if next_timestamp > 0 {
		params["next_timestamp"] = next_timestamp
	}

	var result GetNotificationsResponse
	err := this.callMap("get_notifications", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Fetches the topics from the inbox and returns them along with the number
// of unread topics in the inbox.
// An error is returned if anything (network, auth, jsonrpc) goes wrong.
func (this *Client) ListInboxTopics() (*ListTopicResponse, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["archived"] = 0

	var result ListTopicResponse
	err := this.callMap("topics_list", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *Client) ListArchivedTopics() (*ListTopicResponse, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["archived"] = 1

	var result ListTopicResponse
	err := this.callMap("topics_list", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Performs a search and returns all topics matching the given search filter.
func (this *Client) SearchTopics(filter string) (*ListTopicResponse, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["filter"] = filter

	var result ListTopicResponse
	err := this.callMap("topics_search", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *Client) GetTopic(topicId string) (*Topic, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["id"] = topicId

	var result Topic
	err := this.callMap("topic_get_details", params, &result)
	return &result, err
}

// Add a user to a topic
// Triggers a notification for all sessions of all users of this topic which are online (except for the current session)
// Triggers a message for all other users
func (this *Client) AddTopicReader(topicId string, contactId int) error {
	var params map[string]interface{} = make(map[string]interface{})
	params["topic_id"] = topicId
	params["contact_id"] = contactId

	var result string // We don't care
	return this.callMap("topic_add_user", params, &result)
}

// Creates a new topic with an empty root post (id: 1)
func (this *Client) CreateTopic(topicId string) error {
	var params map[string]interface{} = make(map[string]interface{})
	params["id"] = topicId

	var result string // We don't care
	return this.callMap("topics_create", params, &result)
}

// Returns true if the user was added, false otherwise
func (this *Client) AddContact(email string) (bool, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["contact_email"] = email

	var result bool // We don't care
	err := this.callMap("user_add_contact", params, &result)
	return result, err
}

func (this *Client) GetContacts() ([]User, error) {
	var params map[string]interface{} = make(map[string]interface{})

	var result []User
	err := this.callMap("user_get_contacts", params, &result)
	return result, err
}

func (this *Client) CreatePost(topicId, postId, parentPostId string, intendedReply bool) (bool, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["topic_id"] = topicId
	params["post_id"] = postId
	params["parent_post_id"] = parentPostId
	if intendedReply {
		params["intended_reply"] = 1
	} else {
		params["intended_reply"] = 0
	}

	var result bool
	err := this.callMap("post_create", params, &result)
	return result, err
}

func (this *Client) EditPost(topicId, postId, content string, revisionNo int) (*EditPostResponse, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["topic_id"] = topicId
	params["post_id"] = postId
	params["content"] = content
	params["revision_no"] = revisionNo

	var result EditPostResponse
	err := this.callMap("post_edit", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (this *Client) DeletePost(topicId, postId string) (bool, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["topic_id"] = topicId
	params["post_id"] = postId

	var result bool
	err := this.callMap("post_delete", params, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (this *Client) ChangePostRead(topicId, postId string, read bool) error {
	var params map[string]interface{} = make(map[string]interface{})
	params["topic_id"] = topicId
	params["post_id"] = postId
	params["read"] = 1
	if !read {
		params["read"] = 0
	}

	var result bool // Ignore
	return this.callMap("post_change_read", params, &result)
}

func (this *Client) ChangePostLock(topicId, postId string, lock bool) (bool, error) {
	var params map[string]interface{} = make(map[string]interface{})
	params["topic_id"] = topicId
	params["post_id"] = postId
	params["lock"] = 1
	if !lock {
		params["lock"] = 0
	}

	var result bool
	err := this.callMap("post_change_lock", params, &result)
	return result, err
}

// Same as call(), but remarshals the result into the given result object
func (this *Client) callMap(method string, params map[string]interface{}, result interface{}) error {
	response, err := this.call(method, params)
	if err != nil {
		return err
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	//log.Println(string(data))
	return json.Unmarshal(data, result)
}

func (this *Client) call(method string, params map[string]interface{}) (interface{}, error) {
	id := this.idSeq
	this.idSeq++
	log.Printf("#%04d - Calling %s with %+v\n", id, method, params)
	if this.sessionId != "" {
		params["apikey"] = this.sessionId
	}
	return JsonRpcCall(this.endpoint, method, id, params)
}
