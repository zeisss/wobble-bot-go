package api

// See wobble.api.Client#Subscribe() to get started with notifications
type Subscription struct {
	notifications chan *Notification
	errors        chan error
	stop          chan int
}

func newSubscription() *Subscription {
	var sub Subscription

	sub.notifications = make(chan *Notification)
	sub.errors = make(chan error)
	sub.stop = make(chan int)

	return &sub
}

func (this *Subscription) loop(client *Client) {
	var next_timestamp float64 = -1

	for {
		// If the get_notifications call fails, die
		result, err := client.getNotifications(next_timestamp) // Blocks a long time if no message is on the server!
		if err != nil {
			this.errors <- err
			return
		}

		// Otherwise, write all messages to the channel
		next_timestamp = result.NextTimestamp
		for _, no := range result.Messages {
			this.notifications <- &no // Blocking!
		}

		// Check if we received a signal to exit
		select {
		case <-this.stop:
			return
		default:
		}
	}
}

func (this *Subscription) GetNextNotification() (*Notification, error) {
	select {
	case not := <-this.notifications:
		return not, nil
	case err := <-this.errors:
		return nil, err
	}
	return nil, nil
}
func (this *Subscription) Stop() {
	this.stop <- 1
}
