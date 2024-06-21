package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	session, _     = discordgo.New("Bot ")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")

	integerOptionMinValue = 100.0

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"auto-purger": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			slog.Info("Start deleting messages!")

			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Start deleting messages. Have a nice day!",
				},
			})

			amount := i.Interaction.ApplicationCommandData().Options[0].IntValue() / 100

			for j := int64(0); j < amount; j++ {
				messages, messageErr := s.ChannelMessages(i.ChannelID, 100, "", "", "")

				if messageErr != nil {
					slog.Error(messageErr.Error())
				}

				messagesId := make([]string, 0)

				for _, message := range messages {
					messagesId = append(messagesId, message.ID)
				}

				err := session.ChannelMessagesBulkDelete(i.ChannelID, messagesId)

				if err != nil {
					slog.Error(err.Error())
				}
			}
		},
	}

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "auto-purger",
			Description: "Purge chat",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "amount",
					Description: "Number of messages that will be deleted!",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
					MinValue:    &integerOptionMinValue,
					MaxValue:    1000,
				},
			},
		},
	}
)

func init() {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})
}

func main() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info(fmt.Sprintf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator))
	})

	err := session.Open()

	if err != nil {
		slog.Error(err.Error())
	}

	guilds := session.State.Guilds

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	for _, guild := range guilds {
		for i, command := range commands {
			cmd, err := session.ApplicationCommandCreate(session.State.User.ID, guild.ID, command)
			if err != nil {
				slog.Info(fmt.Sprintf("Cannot create '%v' command: %v", command.Name, err))
			}
			registeredCommands[i] = cmd
		}
	}

	slog.Info("Adding commands...")

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	defer func(session *discordgo.Session) {
		err := session.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(session)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if *RemoveCommands {
		slog.Info("Removing commands...")
		for _, guild := range guilds {
			for _, command := range registeredCommands {
				err := session.ApplicationCommandDelete(session.State.User.ID, guild.ID, command.ID)
				if err != nil {
					slog.Info(fmt.Sprintf("Cannot delete '%v' command: %v", command.Name, err))
				}
			}
		}
	}

	slog.Info("Gracefully shutting down.")
}
