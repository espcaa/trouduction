package main

import "github.com/slack-go/slack"

func (b *Bot) IsBotInChannel(channelID string) (bool, error) {
	channel, err := b.State.SlackClient.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID:     channelID,
		IncludeLocale: false,
	})
	if err != nil {
		return false, err
	}
	return channel.IsMember, nil
}

func (b *Bot) isValidEmoji(emoji string) bool {
	if len(emoji) < 3 || emoji[0] != ':' || emoji[len(emoji)-1] != ':' {
		return false
	}
	return true
}
