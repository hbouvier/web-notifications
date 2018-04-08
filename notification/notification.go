package notification

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/hbouvier/web-notifications/storage"
)

type data struct {
	Href string `json:"href"`
}

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon"`
	Badge string `json:"badge"`
	Data  data   `json:"data"`
}

type Notification struct {
	Subscriber string  `json:"subscriber"`
	Event      Payload `json:"event"`
}

func (notification *Notification) Push(subscriptionJSON string, vapid *storage.VAPID) error {
	// Decode subscription
	subscription := webpush.Subscription{}
	if err := json.NewDecoder(bytes.NewBufferString(subscriptionJSON)).Decode(&subscription); err != nil {
		return err
	}
	jsonString, err := json.Marshal(&notification.Event)
	if err != nil {
		return err
	}
	// Push the notification
	log.Printf("Event: %s", jsonString)
	_, err = webpush.SendNotification([]byte(jsonString), &subscription, &webpush.Options{
		Subscriber:      notification.Subscriber,
		VAPIDPrivateKey: vapid.Private,
	})
	return err
}
