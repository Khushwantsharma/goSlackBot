package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const token string = "xoxb-3679713085909-3682892543011-MfEz6zaYy7tx47F04NgbDdMK"
const appToken string = "xapp-1-A03KN3DFY2K-3711330104880-5e087222cb9a02858eae1cb95aacbc43bdb19cd11a58d0bfb4dbb8f3de7026b5"
const channelID string = "C03LSBH7XL0"

// This will take an event and handle it properly based on the Event Type
func handleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	log.Println("handleEventMessage Invoked")
	switch event.Type {
	// First we check if this is an CallbackEvent
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			err := handleAppMentionEvent(ev, client)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Unsupported event type")
	}
	return nil
}

// handleAppMentionEvent is used to take care of the AppMentionEvent when the bot is mentioned
func handleAppMentionEvent(event *slackevents.AppMentionEvent, client *slack.Client) error {
	log.Println("handleAppMentionEvent Invoked")
	// Grab the user name based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}
	// Check if the user said Hello to the bot
	text := strings.ToLower(event.Text)

	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}
	// Add Some default context like user who mentioned the bot
	attachment.Fields = []slack.AttachmentField{}
	attachment.Color = "#4af030"
	log.Printf("User Message: %s and UserName: %s\n", text, user.Name)
	if strings.Contains(text, "hello") {
		//To Greet User
		attachment.Text = fmt.Sprintf("Hello %s", user.Name)
		attachment.Pretext = "Greetings"
	} else {
		// Send a custamize message
		attachment.Text = userinput()
		attachment.Pretext = "Service"
	}
	// Send the message to the channel
	// The Channel is available in the event message
	_, _, err = client.PostMessage(channelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

//To make custamize input for Bot .
func userinput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Message For User:")
	input, _ := reader.ReadString('\n')
	return input

}

func main() {

	// Create a new client to slack by giving token
	// Set debug to true while developing
	// Also add a ApplicationToken option to the client
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(false),
		// Option to set a custom logger
		// socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	go func(client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			case event := <-socketClient.Events:
				// We have a new Events, let's type switch the event
				// Add more use cases here if you want to listen to other events.
				switch event.Type {
				// handle EventAPI events
				case socketmode.EventTypeEventsAPI:

					// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					// We need to send an Acknowledge to the slack server that we recevied the message
					socketClient.Ack(*event.Request)

					// log.Println(eventsAPIEvent)
					// Now we have an Events API event
					err := handleEventMessage(eventsAPIEvent, client)
					if err != nil {
						// Replace with actual err handeling
						log.Fatal(err)
					}
				}

			}
		}
	}(client, socketClient)

	socketClient.Run()
}
