package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	api "github.com/ZeissS/wobble-go-client"
)

func Connect(endpoint, email, password string) (*api.Client, error) {
	client := api.NewClient(endpoint)
	if err := client.Login(email, password); err != nil {
		return nil, err
	}
	return client, nil
}

func runBot(client *api.Client) {
	currentUser, err := client.GetCurrentUser()
	if err != nil {
		log.Fatal("Failed to fetch current user: " + err.Error())
	}
	log.Println(currentUser)

	result, err := client.SearchTopics("root")
	if err != nil {
		log.Fatal("Listing Inbox Topics failed: ", err)
	}

	fmt.Printf("You have %d unread topics\n", result.InboxUnreadTopics)

	for _, t := range result.Topics {
		fmt.Println(t)

		topic, err := client.GetTopic(t.TopicId)
		if err != nil {
			log.Println("Failed to fetch topic ", t.TopicId, ": "+err.Error())
			continue
		}

		postId := fmt.Sprintf("%s-gobot-%d", topic.TopicId, os.Getpid())
		if _, err := client.CreatePost(t.TopicId, postId, "1", true); err != nil {
			log.Println("Failed to create post:", err)
			continue
		}
		if _, err := client.EditPost(t.TopicId, postId, "Hello World from GoBot", 1); err != nil {
			log.Println("Failed to edit post:", err)
			continue
		}
		if _, err := client.DeletePost(t.TopicId, postId); err != nil {
			log.Println("Failed to delete post:", err)
		}
	}
}

func findContactByEmail(client *api.Client, email string) int {
	var users []api.User
	users, err := client.GetContacts()
	if err != nil {
		return -2
	}

	for _, user := range users {
		if user.Email == email {
			return user.UserId
		}
	}
	return -1
}

// This will enter a long running loop to talk to someone ;)
// 
func runTalkBot(client *api.Client, otherUser string) {
	var talkTopicId = "c68c1a28-7c5a-11e2-844f-68a86d44bfa4"

	_, err := client.GetTopic(talkTopicId)

	if err != nil {
		if apierr, ok := err.(api.WobbleApiError); !ok || apierr.Message != "Illegal Access!" {
			log.Fatal("Failed to get topic", err)
		}

		// We blindly assume this error means, that the topic does not exist yet. So we create it
		if err := client.CreateTopic(talkTopicId); err != nil {
			log.Fatal("Failed to create talk topic", err)
		}

		client.AddContact(otherUser)

		otherUserId := findContactByEmail(client, otherUser)
		if otherUserId <= 0 {
			log.Fatal("Failed to find user "+otherUser, ":", otherUserId)
		}
		client.AddTopicReader(talkTopicId, otherUserId)
		client.EditPost(talkTopicId, "1", "Hi "+otherUser, 1) // Edit the root post
	}

	time.Sleep(1 * time.Second)

	currentUserId, err := client.GetCurrentUserId()
	if err != nil {
		log.Fatal("Failed to read current user", err)
	}

	time.Sleep(1 * time.Second)

	checkTopic(client, currentUserId, talkTopicId)

	subscription := client.SubscribeNotifications()

	for {
		notification, err := subscription.GetNextNotification()
		if err != nil {
			log.Fatal(err)
			break
		}
		log.Println("Notification", notification)

		if notification.TopicId == talkTopicId {
			checkTopic(client, currentUserId, talkTopicId)
			time.Sleep(1 * time.Second)
		}
	}
	defer subscription.Stop()
}

type UserList []int

func (users UserList) contains(userId int) bool {
	for _, user := range users {
		if user == userId {
			return true
		}
	}
	return false
}

// Checks if there is the need to respond to anything in the given topic
// If so, the response is done ;)
func checkTopic(client *api.Client, currentUserId int, talkTopicId string) {
	var responses []string = []string{
		"What do you mean?",
		"I don't get you.",
		"Do you know <a href=\"http://devopsreactions.tumblr.com/\">DevOps Reactions</a>?",
		"Oh come ON!",
		"I'll be back!",
		"Kill me, NOW!",
	}

	topic, err := client.GetTopic(talkTopicId)
	if err != nil {
		log.Fatal("Failed to read topic: ", err)
	}

	for _, post := range topic.Posts {
		if post.Deleted == 1 || post.Unread == 0 || post.Lock != nil ||
			UserList(post.Users).contains(currentUserId) {
			continue
		}
		// It is none of my posts and it is unread
		log.Println("Replying to post ", post.PostId+" from user ", post.Users)
		_, err := client.CreatePost(talkTopicId, post.PostId+"1", post.PostId, true)
		if err != nil {
			log.Println("Failed to create post", err)
		} else {
			client.EditPost(talkTopicId, post.PostId+"1", responses[rand.Intn(len(responses))], ^post.IntendedPost&1)
		}
		client.ChangePostRead(talkTopicId, post.PostId, true) // Mark post as read

		time.Sleep(1 * time.Second)
	}

}

func parseArguments() (string, string, string, bool) {
	endpoint := flag.String("endpoint", "http://wobble.moinz.de/api/endpoint.php", "The HTTP endpoint to communicate with")
	email := flag.String("email", "", "Email for wobble api")
	password := flag.String("password", "", "The password")
	showHelp := flag.Bool("help", false, "Shows the help")
	flag.BoolVar(showHelp, "h", false, "Shows the help")

	flag.Parse()

	if *showHelp || flag.NArg() > 0 {
		fmt.Println("ERROR: Unknown arguments.")
		flag.PrintDefaults()
		return "", "", "", false
	}

	return *endpoint, *email, *password, true
}

func main() {
	rand.Seed(123414123)
	fmt.Println("# Wobble Bot v0.1")

	endpoint, username, password, ok := parseArguments()

	if !ok {
		return
	}

	client, err := Connect(endpoint, username, password)
	if err != nil {
		log.Fatal("Failed to login: " + err.Error())
	}
	if version, err := client.WobbleVersion(); err != nil {
		log.Fatal("Failed to fetch wobble version", err)
	} else {
		log.Println("# Wobble Server Version", version)
	}
	defer client.Logout()

	// runBot(client)
	runTalkBot(client, "stephan.zeissler@moinz.de")
}
