package main

import (
	"encoding/json"
	"log"

	_ "embed"

	"github.com/espcaa/trouduction/ai"
	"github.com/google/shlex"
	"github.com/slack-go/slack"
)

//go:embed prompt.txt
var aiSystemPrompt string

type GptJson struct {
	Options []string `json:"options"`
	Ok      bool     `json:"ok"`
}

func SendResponseURLMessage(url string, blocks *slack.Blocks, ephemeral bool, replaceOriginal bool) error {
	response := slack.WebhookMessage{
		Blocks:          blocks,
		ResponseType:    "ephemeral",
		ReplaceOriginal: replaceOriginal,
	}
	if !ephemeral {
		response.ResponseType = "in_channel"
	}

	err := slack.PostWebhook(url, &response)
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
		err := SendResponseURLMessage(command.ResponseURL, &blocks, true, false)
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
		err := SendResponseURLMessage(command.ResponseURL, &blocks, true, false)

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
		err := SendResponseURLMessage(command.ResponseURL, &blocks, true, false)

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
	err = SendResponseURLMessage(command.ResponseURL, &loadMessageBlocks, true, false)
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

	// try to parse it
	var gptJson GptJson
	err = json.Unmarshal([]byte(responseContent), &gptJson)
	if err != nil {
		log.Printf("error parsing ai response: %v", err)
		// log the response content for debugging
		log.Printf("ai response content: %s", responseContent)
		// send a message to the user saying that the ai response was invalid
		blocks := slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "uh looks like our clanker had a stroke. try again. "+err.Error(), false, false),
					nil,
					nil,
				),
			},
		}
		err := SendResponseURLMessage(command.ResponseURL, &blocks, true, false)
		if err != nil {
			log.Printf("error sending ephemeral message: %v", err)
		}
		return
	}

	if !gptJson.Ok {
		blocks := slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", "uhm we got a message from our clanker saying that this emoji is not trouductable. try again if you think this is a mistake.", false, false),
					nil,
					nil,
				),
			},
		}
		err := SendResponseURLMessage(command.ResponseURL, &blocks, true, false)
		if err != nil {
			log.Printf("error sending ephemeral message: %v", err)
		}
		return
	}

	// now we have choices :3
	// send a message to the user with the choices

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					":blobhaj_party: here are some possible trouductions for `"+emoji+"` "+emoji+" !!",
					false, false),
				nil, nil,
			),
			slack.NewDividerBlock(),
		},
	}

	for _, option := range gptJson.Options {
		button := slack.NewButtonBlockElement(
			"emoji&"+StripColons(emoji)+"&"+option,
			option, // put the value in Value, not the action_id
			slack.NewTextBlockObject("plain_text", "create", false, false),
		)
		button.Style = slack.StylePrimary

		blocks.BlockSet = append(blocks.BlockSet,
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "*"+option+"*", false, false),
				nil,
				slack.NewAccessory(button),
			),
		)
	}

	// add a cancel button at the end
	cancelButton := slack.NewButtonBlockElement(
		"cancel&"+StripColons(emoji),
		"cancel",
		slack.NewTextBlockObject("plain_text", "cancel", false, false),
	)
	cancelButton.Style = slack.StyleDanger

	blocks.BlockSet = append(blocks.BlockSet,
		slack.NewDividerBlock(),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "or cancel this operation if you don't want to add any of these.", false, false),
			nil,
			slack.NewAccessory(cancelButton),
		),
	)

	err = SendResponseURLMessage(command.ResponseURL, &blocks, true, false)
	if err != nil {
		log.Printf("error sending ephemeral message: %v", err)
	}

	// done now?
}

func StripColons(emoji string) string {
	if len(emoji) < 2 {
		return emoji
	}
	if emoji[0] == ':' && emoji[len(emoji)-1] == ':' {
		return emoji[1 : len(emoji)-1]
	}
	return emoji
}
