package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/omer1998/chat-app-go.git/chat/app/sdk/chat"

	"github.com/rivo/tview"
)

type App struct {
	TvApp   *tview.Application
	TvFlex  *tview.Flex
	TvMsgs  *tview.TextView
	TvUsers *tview.List
	Client  *Client
	Config  *Config
}

func NewApp(client *Client, config *Config) *App {
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
	usersList.SetTitle("Contacts")
	// here we update the user/ contact list according to the contacts of this user
	for index, usr := range client.config.Contacts {
		usersList.AddItem(usr.Name, usr.Id, rune(index+49), nil)
	}

	// here we need to read and show the messages that was sent and recieved with this selected user
	usersList.SetChangedFunc(func(index int, name, id string, shortcut rune) {

		usr, err := config.LookUpUser(id)
		if err != nil {
			fmt.Fprintf(msgs, "\nsystem: %s", err.Error())
		}
		msgs.Clear()
		for _, msg := range usr.Messages {
			fmt.Fprintf(msgs, "%s", msg)

		}
		// usersList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		// 	usersList.SetItemText(i, )
		// })
		usersList.SetItemText(index, usr.Name, id)

	})

	usersList.SetBorder(true)
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
			err := client.Send(chat.InMessage{ToId: id, Msg: newMsg})
			if err != nil {
				fmt.Fprintf(msgs, "\nsystem: %s", err.Error())
			} else {
				config.AddMessage(id, fmt.Sprintf("\nyou: %s", newMsg))
				fmt.Fprintf(msgs, "\nyou: %s", newMsg)
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

	return &App{TvApp: app, TvFlex: flex, TvMsgs: msgs, TvUsers: usersList, Client: client, Config: config}
}

type uiMsgHandler func(from string, msg string)
type uiUpdateUsers func(usr user)

func (a *App) Run() error {

	go a.Client.Handshake(a.WriteText, a.UpdateUsers)

	// if err != nil {
	// 	a.WriteText("system: error reading incoming msgs" + err.Error())
	// }
	return a.TvApp.SetRoot(a.TvFlex, true).Run()
}

func (a *App) WriteText(from string, text string) {
	currentUserIdx := a.TvUsers.GetCurrentItem()
	name, _ := a.TvUsers.GetItemText(currentUserIdx)
	if name == from {
		//mean the selcted user is the same user who send this message so we display the message directly
		fmt.Fprintf(a.TvMsgs, "\n------")
		fmt.Fprintf(a.TvMsgs, "\n%s: %s", from, text)
	} else {
		//if the selected user on screen (terminal) is not the user who send this message we need to modify the name of the user who send this message
		// in order to indicate there is a new message from this usere
		for _, usr := range a.Config.Contacts {
			if usr.Name == from {
				fromUserIndex := a.TvUsers.FindItems(usr.Name, usr.Id, false, true)[0]
				fromName, fromId := a.TvUsers.GetItemText(fromUserIndex)
				a.TvUsers.SetItemText(fromUserIndex, fmt.Sprintf("* %s", fromName), fromId)
				a.TvApp.Draw()

			}
		}

	}

}

func (a *App) UpdateUsers(usr user) {
	usersNum := a.TvUsers.GetItemCount()
	a.TvUsers.AddItem(usr.Name, usr.Id, rune(usersNum+49), nil)
}

// func (a *App) getToUser() chatapp.User {
// 	currentItem := a.TvUsers.GetCurrentItem()
// 	usrName, id := a.TvUsers.GetItemText(currentItem)
// 	return chatapp.User{
// 		Name: usrName,
// 		Id:   uuid.MustParse(id),
// 	}
// }
