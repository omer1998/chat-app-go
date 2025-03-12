package chat

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"

	"github.com/rivo/tview"
)

type App struct {
	TvApp   *tview.Application
	TvFlex  *tview.Flex
	TvMsgs  *tview.TextView
	TvUsers *tview.List
	Client  *Client
}

func NewApp(client *Client) *App {
	app := tview.NewApplication()
	app.EnableMouse(true)

	// =======================================================
	msgs := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	msgs.SetBorder(true)
	// fmt.Fprintf(msgs, "%s", "This is omer faris")

	// =======================================================
	usersList := tview.NewList()
	usersList.AddItem("omer faris", "1b0deec0-5c9d-42bd-98ee-59e39d5a8105", 'a', nil)
	usersList.AddItem("ahmed faris", "6913f142-ffe2-49c7-a222-5a1b1638992b", 'b', nil)
	// =======================================================
	msgInput := tview.NewTextArea().SetPlaceholder("Enter your message here")
	// msgInput.SetBorder(true)

	// =======================================================
	submitBtn := tview.NewButton("submit")
	submitBtn.SetRect(0, 0, 20, 6)
	submitBtn.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			// msgs.Clear()
			newMsg := msgInput.GetText()
			msgInput.SetText("", false)
			// here we need to send msg (inmessage) from me to a specified user (this specified user will be determine by the selects user from user list view)
			currentItem := usersList.GetCurrentItem()
			_, id := usersList.GetItemText(currentItem)
			// toUser := chatapp.User{Id: uuid.MustParse(id), Name: usrName}
			err := client.Send(chat.InMessage{ToId: uuid.MustParse(id), Msg: newMsg})
			if err != nil {
				fmt.Fprintf(msgs, "\nsystem: %s", err.Error())
			} else {
				fmt.Fprintf(msgs, "you: %s\n", newMsg)
			}

		}
		return action, event
	})
	// submitBtn.SetBorder(true)

	// =======================================================

	flex := tview.NewFlex()
	// flex.SetDirection(tview.FlexRow)
	flex.
		AddItem(usersList, 0, 2, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(msgs, 0, 9, true).
				AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
					AddItem(msgInput, 0, 9, true).
					AddItem(submitBtn, 0, 1, true), 0, 1, false), 0, 10, false)

	return &App{TvApp: app, TvFlex: flex, TvMsgs: msgs, TvUsers: usersList, Client: client}
}

type messageHandler func(from string, msg string)

func (a *App) Run() error {

	go a.Client.Handshake(a.WriteText)

	// if err != nil {
	// 	a.WriteText("system: error reading incoming msgs" + err.Error())
	// }
	return a.TvApp.SetRoot(a.TvFlex, true).Run()
}

func (a *App) WriteText(from string, text string) {
	fmt.Fprintf(a.TvMsgs, "%s: %s\n", from, text)

}

// func (a *App) getToUser() chatapp.User {
// 	currentItem := a.TvUsers.GetCurrentItem()
// 	usrName, id := a.TvUsers.GetItemText(currentItem)
// 	return chatapp.User{
// 		Name: usrName,
// 		Id:   uuid.MustParse(id),
// 	}
// }
