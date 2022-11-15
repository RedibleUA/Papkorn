package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"main/DB"
	"os"
	"os/signal"
)

const (
	TOKEN = "TOKEN"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", TOKEN, "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "list-films",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "List films",
		},
		{
			Name:        "add-film",
			Description: "Add film",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title-name",
					Description: "Name of Film",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "URL of films",
					Required:    false,
				},
			},
		},
		{
			Name:        "random-film",
			Description: "Random film",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"list-films": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

			FilmDict := DB.ListFilms()

			count := 1
			textList := ""

			for key, value := range FilmDict {
				textList += fmt.Sprintf("**%d.** [%s](%s)\n", count, key, value)
				count++
			}

			title := "Список"
			if textList == "" {
				title += " пуст"
				textList = "Здесь будут фильмы"
			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       title,
							Description: textList,
							Color:       15750227,
							Thumbnail: &discordgo.MessageEmbedThumbnail{
								URL: "https://images.vexels.com/media/users/3/261537/isolated/lists/6fc20982097cd6d7c96bc5a9e0604c88-popcorn-red-icon.png",
							},
						},
					},
				},
			})
			if err != nil {
				return
			}
		},
		"add-film": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Access options in the order provided by the user.
			options := i.ApplicationCommandData().Options

			// Or convert the slice into a map
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			answer := ""
			if _, ok := optionMap["url"]; ok {
				answer = DB.AddFilm(optionMap["title-name"].StringValue(), optionMap["url"].StringValue())
			} else {
				answer = DB.AddFilm(optionMap["title-name"].StringValue(), "")
			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: answer,
				},
			})
			if err != nil {
				return
			}

		},
		"random-film": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

			name, url := DB.RandomFilm()
			data := &discordgo.InteractionResponseData{}
			if name != "" {

				// https://github.com/bwmarrin/discordgo/pull/945/files
				// https://github.com/bwmarrin/discordgo/wiki/FAQ#sending-embeds
				embed := []*discordgo.MessageEmbed{
					{
						Title:       fmt.Sprintf("Выпал: %s", name),
						Description: "Повезло, повезло)\nПриятного просмотра",
						Color:       15750227,
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: "https://images.vexels.com/media/users/3/261537/isolated/lists/6fc20982097cd6d7c96bc5a9e0604c88-popcorn-red-icon.png",
						},
					},
				}

				if url != "" {

					// https://git.mgmcomp.net/thisnthat/discordgo/src/6f37fbe58ad70ba373806b3b7901da198e28eec9/examples/components/main.go
					button := []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label: "Смотреть",
									Style: discordgo.LinkButton,
									URL:   url,
								},
							},
						},
					}

					data = &discordgo.InteractionResponseData{
						Embeds:     embed,
						Components: button,
					}
				} else {
					data = &discordgo.InteractionResponseData{
						Embeds: embed,
					}
				}

			} else {
				embed := []*discordgo.MessageEmbed{
					{
						Title:       "Фильмов нету",
						Description: "Добавь фильмы и крути)",
						Color:       15750227,
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: "https://images.vexels.com/media/users/3/261537/isolated/lists/6fc20982097cd6d7c96bc5a9e0604c88-popcorn-red-icon.png",
						},
					},
				}

				data = &discordgo.InteractionResponseData{
					Embeds: embed,
				}

			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: data,
			})
			if err != nil {
				return
			}

		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer func(s *discordgo.Session) {
		err := s.Close()
		if err != nil {

		}
	}(s)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		// We need to fetch the commands, since deleting requires the command ID.
		// We are doing this from the returned commands on line 375, because using
		// this will delete all the commands, which might not be desirable, so we
		// are deleting only the commands that we added.
		registeredCommands, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		if err != nil {
			log.Fatalf("Could not fetch registered commands: %v", err)
		}

		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
