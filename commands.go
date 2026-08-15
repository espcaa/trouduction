package main

import (
	"log"

	_ "embed"

	"github.com/espcaa/trouduction/ai"
	"github.com/google/shlex"
	"github.com/slack-go/slack"
)

//go:embed prompt.txt
var aiSystemPrompt string

func SendResponseURLMessage(command slack.SlashCommand, blocks *slack.Blocks, ephemeral bool) error {
	response := slack.WebhookMessage{
		Blocks:       blocks,
		ResponseType: "ephemeral",
	}
	if !ephemeral {
		response.ResponseType = "in_channel"
	}

	err := slack.PostWebhook(command.ResponseURL, &response)
	return err
}

func (b *Bot) handleTrouductionCommand(command slack.SlashCommand) {
	// check if an emoji was given

	args, err := shlex.Split(command.Text)
	if err != nil {
		blocks := slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "nauurr i couldn't parse arguments, wtf did u do: "+err.Error(), false, false),
					nil,
					nil,
				),
			},
		}
		err := SendResponseURLMessage(command, &blocks, true)
		if err != nil {
			log.Printf("error sending ephemeral message: %v", err)
		}
		return
	}
	if len(args) == 0 {

		blocks := slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "please give us an emoji mmmh (usage `/trouduction :emojiname:`)", false, false),
					nil,
					nil,
				),
			},
		}
		err := SendResponseURLMessage(command, &blocks, true)

		if err != nil {
			log.Printf("error sending ephemeral message: %v", err)
		}

		return
	}

	emoji := args[0]
	if !b.isValidEmoji(emoji) {

		blocks := slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "this isn't a valid emoji :sob-wx:", false, false),
					nil,
					nil,
				),
			},
		}
		err := SendResponseURLMessage(command, &blocks, true)

		if err != nil {
			log.Printf("error sending ephemeral message: %v", err)
		}

		return
	}

	loadMessageBlocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", ":mc_loading: Trouducting `"+emoji+"`", false, false),
				nil,
				nil,
			),
		},
	}
	err = SendResponseURLMessage(command, &loadMessageBlocks, true)
	if err != nil {
		log.Printf("error sending ephemeral message: %v", err)
		return
	}

	// once we sent that conformation, we should deal with the actual thing

	var aiConversation []ai.AiMessage = []ai.AiMessage{
		{
			Role:    "system",
			Content: aiSystemPrompt,
		},
		{
			Role:    "user",
			Content: emoji,
		},
	}

	aiResponse, err := b.State.AiClient.Complete(aiConversation)
	if err != nil {
		log.Printf("error getting ai response: %v", err)
		return
	}

	responseContent := aiResponse.GetContent()

	responseBlocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", responseContent, false, false),
				nil,
				nil,
			),
		},
	}

	err = SendResponseURLMessage(command, &responseBlocks, false)
	if err != nil {
		log.Printf("error sending in channel message: %v", err)
	}
}
