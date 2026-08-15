package main

import (
	"log"
	"os"

	"github.com/espcaa/trouduction/ai"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type Config struct {
	SlackBotToken  string
	SlackAppToken  string
	SlackUserToken string
	AiApiKey       string
	AiModel        string
	AiBaseUrl      string
}

type BotState struct {
	AiClient     *ai.AiClient
	SlackClient  *slack.Client
	SocketClient *socketmode.Client
}

func NewBotState(config Config) *BotState {
	aiClient := ai.NewAiClient(config.AiBaseUrl, config.AiApiKey, config.AiModel)
	slackClient := slack.New(config.SlackBotToken, slack.OptionAppLevelToken(config.SlackAppToken))
	socketClient := socketmode.New(slackClient)

	return &BotState{
		AiClient:     aiClient,
		SlackClient:  slackClient,
		SocketClient: socketClient,
	}
}

func loadConfig() Config {
	godotenv.Load(".env")
	SlackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	SlackAppToken := os.Getenv("SLACK_APP_TOKEN")
	SlackUserToken := os.Getenv("SLACK_USER_TOKEN")
	AiApiKey := os.Getenv("AI_API_KEY")
	AiModel := os.Getenv("AI_MODEL")
	AiBaseUrl := os.Getenv("AI_BASE_URL")
	return Config{
		SlackBotToken:  SlackBotToken,
		SlackAppToken:  SlackAppToken,
		SlackUserToken: SlackUserToken,
		AiApiKey:       AiApiKey,
		AiModel:        AiModel,
		AiBaseUrl:      AiBaseUrl,
	}
}

func main() {
	config := loadConfig()
	botState := NewBotState(config)
	bot := NewBot(botState)
	bot.Start()
}

type Bot struct {
	State *BotState
}

func NewBot(state *BotState) *Bot {
	return &Bot{
		State: state,
	}
}

func (b *Bot) Start() {
	log.Println("Bot started...")

	handler := socketmode.NewSocketmodeHandler(b.State.SocketClient)

	// handle all slash commands
	handler.Handle(socketmode.EventTypeSlashCommand, b.handleSlashCommand)

	if err := handler.RunEventLoop(); err != nil {
		log.Fatalf("socket mode run error: %v", err)
	}
}

func (b *Bot) handleSlashCommand(evt *socketmode.Event, client *socketmode.Client) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		return
	}

	// always ack instantly
	client.Ack(*evt.Request)

	log.Printf("slash command %q from user %s: %q", cmd.Command, cmd.UserID, cmd.Text)

	// check if the bot is in this channel
	// not needed rn as we're using the response url for everything
	// inChannel, err := b.IsBotInChannel(cmd.ChannelID)

	// handle every slack command
	switch cmd.Command {
	case "/trouduction":
		b.handleTrouductionCommand(cmd)
	default:
		client.PostMessage(cmd.ChannelID, slack.MsgOptionText("what is this command.", false))
	}
}
