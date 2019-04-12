/*
Copyright 2019 tira

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * selfbot.go contains the main funtions of the self bot.
 */
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	customCommands = make(map[string]string)
)

const (
	COLOR_ERROR   = 0x8b1117 /* Dark red */
	COLOR_SUCCESS = 0x006606 /* Dark green */
	COLOR_NOTICE  = 0x00549d /* Dark blue */
)

func Selfbot(Discord *discordgo.Session) {
	/*
	 * ParseCustomCommands() uses json.Unmashal to pass the data of commands.json
	 * directly to the customCommands map.2
	 */
	ParseCustomCommands()
	Discord.AddHandler(OnMessageCreate)
	err := Discord.Open()
	if err != nil {
		Fatal("Failed to start selfbot, " + err.Error())
	}
	fmt.Println("Using prefix", UserConfig.SelfBotPrefix)
	fmt.Println("Listening for messages")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt,
		os.Kill)
	<-signalChan
}

func OnMessageCreate(Discord *discordgo.Session,
	message *discordgo.MessageCreate) {
	if message.Author.ID != Discord.State.User.ID ||
		!strings.HasPrefix(message.Content, UserConfig.SelfBotPrefix) {
		/*
		 * Quickly ignore all messages not send by the bot owner, and messages that
		 * don't start with the bot's identifier.
		 */
		return
	}
	command := strings.Split(
		message.Content[len(UserConfig.SelfBotPrefix):len(message.Content)], " ")[0]
	var messageContent string
	if len(strings.Split(message.Content, " ")) > 1 {
		messageContent = message.Content[len(UserConfig.SelfBotPrefix)+
			len(command)+1 : len(message.Content)]
	}
	newMessage := discordgo.NewMessageEdit(message.ChannelID, message.ID)
	/*
	 * We only want to display the embed here, so we empty out the message
	 * "content".
	 */
	newMessage.SetContent("")
	switch {
	case command == "about":
		newMessage.SetEmbed(&discordgo.MessageEmbed{
			URL:         "https://github.com/t1ra/blacksheep",
			Title:       "Blacksheep",
			Description: "The Discord tooling powerhouse.",
			Color:       COLOR_NOTICE,
		})
	case command == "help":
		newMessage.SetEmbed(&discordgo.MessageEmbed{
			Title:       "Blacksheep",
			Description: "Help",
			Color:       COLOR_NOTICE,
			Fields:      append(HelpFields(), CustomCommands(customCommands)...),
		})
	case command == "huge":
		newMessage.SetContent(Huge(messageContent))
	case command == "copypasta":
		newMessage.SetContent(Copypasta(UserConfig.SelfBotCopypastas))
	case command == "command" && len(strings.Split(messageContent, " ")) > 1:
		switch strings.Split(messageContent, " ")[0] {
		case "new":
			newMessage.SetEmbed(
				NewCustomCommand(messageContent[4:len(messageContent)]))
		case "delete":
			newMessage.SetEmbed(DeleteCustomCommand(strings.Split(
				messageContent, " ")[1]))
		default:
			newMessage.SetEmbed(&discordgo.MessageEmbed{
				Title:       "Hva?",
				Description: "That doesn't look like a valid sub-command.",
				Color:       COLOR_ERROR,
			})
		}
	case command == "owoify":
		newMessage.SetContent(Owoify(messageContent))
	default:
		if content, ok := customCommands[command]; ok {
			newMessage.SetContent(content)
		} else {
			newMessage.SetEmbed(&discordgo.MessageEmbed{
				Title: "Hva?",
				Description: "That doesn't look like a valid command. Try one of" +
					" these:",
				Color:  COLOR_ERROR,
				Fields: HelpFields(),
			})
		}
	}
	/*
	 * Edit the message containing the command, replacing its contents with
	 * whatever was generated by the switch.
	 */
	Discord.ChannelMessageEditComplex(newMessage)
}
