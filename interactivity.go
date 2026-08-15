package main

import (
	"log"
	"strings"

	"github.com/slack-go/slack"
)

func (b *Bot) HandleEmojiAddInteractivity(payload slack.InteractionCallback) {
	action := payload.ActionCallback.BlockActions[0]

	parts := strings.SplitN(action.ActionID, "&", 3)
	if len(parts) < 3 {
		log.Printf("unexpected action id %q: want emoji&<old>&<new>", action.ActionID)
		return
	}
	oldEmoji := parts[1]
	newEmoji := parts[2]

	img, contentType, err := b.GetEmoji(oldEmoji)
	if err != nil {
		log.Printf("error getting emoji image for %q: %v", oldEmoji, err)
		return
	}

	err = b.UploadEmoji(newEmoji, img, contentType)
	if err != nil {
		log.Printf("error uploading emoji %q: %v", newEmoji, err)
		return
	}

	log.Printf("successfully added emoji %q from %q", newEmoji, oldEmoji)

	// delete the original message & send a new message to the user saying the emoji was added
	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn",
					":blobhaj_party: successfully added emoji `"+newEmoji+"` :"+newEmoji+": from `"+oldEmoji+"` :"+oldEmoji+": !!",
					false, false),
				nil, nil,
			),
		},
	}

	err = SendResponseURLMessage(payload.ResponseURL, &blocks, true, true)
	if err != nil {
		log.Printf("error sending ephemeral message: %v", err)
	}

}
