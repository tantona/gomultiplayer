package main

import (
	"os"
	"tantona/gomultiplayer/server/state"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"go.uber.org/zap"
)

func initKeyboard(client *Client, signalChan chan os.Signal) {
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		myData := clientState.GetPlayerData()
		logger.Debug("GetPlayerData", zap.Any("myData", myData))
		switch key.Code {
		case keys.CtrlC, keys.Escape:
			signalChan <- os.Interrupt

			// return true, nil // Return true to stop listener
		case keys.Up:
			msg, _ := createUpdatePlayerMsg(&state.PlayerData{Name: myData.Name, Score: myData.Score + 10})
			logger.Debug("increase score", zap.Any("msg", msg))
			client.Send(msg)
		case keys.Down:
			msg, _ := createUpdatePlayerMsg(&state.PlayerData{Name: myData.Name, Score: myData.Score - 10})
			logger.Debug("decrease score", zap.Any("msg", msg))
			client.Send(msg)
		}

		return false, nil // Return false to continue listening
	})
}
