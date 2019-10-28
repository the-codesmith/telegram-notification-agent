/*
Responsd to ZFS event HTTP requests
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/particleman-smith/telegram-notification-agent/backend/telegram"
)

var telegramBot = telegram.NewBot()

/*
Test sends the given message via the Telegram Bot
*/
func Test(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("ZFS Event 'Test' received.")

	fmt.Println("Sending Telegram message.")
	err := telegramBot.SendMessage("Hey, I just got a test message!")

	returnMsg := ""

	if err != nil {
		returnMsg = "Message failed to send."

		fmt.Println(err.Error())
	} else {
		returnMsg = "Message sent."
	}

	fmt.Println(returnMsg)
	json.NewEncoder(writer).Encode(returnMsg)
}

/*
Error sends an error message via the Telegram Bot. The error message is based on the request URL path.
*/
func Error(writer http.ResponseWriter, request *http.Request) {
	// Read body and handle errors
	body, readErr := ioutil.ReadAll(request.Body)
	if readErr != nil {
		http.Error(writer, readErr.Error(), 500)
		return
	}
	defer request.Body.Close()

	// Parse the body into an interface
	var bodyMap map[string]interface{}
	parseErr := json.Unmarshal([]byte(body), &bodyMap)

	if parseErr != nil {
		http.Error(writer, "Could not parse body.\n"+parseErr.Error(), 500)
		return
	}

	send := true // Whether or not to send the message (body contents may warrant not sending a message)
	msg := "ERROR\n"

	// Error types
	switch request.URL.Path {
	// Bash
	case "/bash-event/exec-failure":
		// TODO: Get more information from request about which script was run (should be sent by the Python script)
		script := bodyMap["script"]
		msg += "I encountered a failure running the Python script `" + fmt.Sprintf("%v", script) + "`."
	// ZFS
	case "/zfs-event/data-error":
		msg += "I noticed data corruption on a ZFS vdev."
	case "/zfs-event/zpool-state":
		// TODO: Use body
		msg += "A disk in a ZFS vdev has gone unavailable!"
	// Backup
	case "/backup-event/failure":
		msg += "I encountered a failure backing up /home, /etc, or /var."
	}

	if send {
		// Send the message via Telegram
		sendErr := telegramBot.SendMessage(msg)

		// Handle Telegram send errors
		if sendErr != nil {
			returnMsg := "Message failed to send."
			fmt.Println(returnMsg)
			json.NewEncoder(writer).Encode(returnMsg)
		}
	}

}
