package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

type EmojiSearchRequest struct {
	Count           int    `json:"count"`
	Query           string `json:"query"`
	Token           string `json:"token"`
	EntrepriseToken string `json:"entreprise_token"`
}

type EmojiSearchResponse struct {
	Ok      bool `json:"ok"`
	Results []struct {
		Name string `json:"name"`
		URL  string `json:"value"`
	} `json:"results"`
}

func (b *Bot) GetEmoji(name string) ([]byte, string, error) {
	url := b.State.Config.SelfBotEdgeUrl + "/emojis/search"

	jsonPayload := EmojiSearchRequest{
		Count:           25,
		Query:           name,
		Token:           b.State.Config.SelfBotToken,
		EntrepriseToken: b.State.Config.SelfBotToken,
	}

	log.Printf("Sending request to %s with payload: %+v", url, jsonPayload)

	payloadBytes, err := json.Marshal(jsonPayload)
	if err != nil {
		return nil, "", errors.New("failed to marshal payload: " + err.Error())
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, "", errors.New("failed to create request: " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "d="+b.State.Config.DCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", errors.New("failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to get emoji: %s", resp.Status)
	}

	var searchResponse EmojiSearchResponse

	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		return nil, "", err
	}

	if !searchResponse.Ok || len(searchResponse.Results) == 0 {
		return nil, "", errors.New("emoji not found, see full response: " + fmt.Sprintf("%+v", searchResponse))
	}

	emojiURL := searchResponse.Results[0].URL

	// Now fetch the emoji image data
	imageResp, err := http.Get(emojiURL)
	if err != nil {
		return nil, "", err
	}
	defer imageResp.Body.Close()

	if imageResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch emoji image: %s", imageResp.Status)
	}

	imageData, err := io.ReadAll(imageResp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := imageResp.Header.Get("Content-Type")

	return imageData, contentType, nil
}

func (b *Bot) UploadEmoji(emoji string, data []byte, contentType string) error {
	apiUrl := b.State.Config.SelfBotUrl + "/api/emoji.add"

	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)

	if err := w.WriteField("token", b.State.Config.SelfBotToken); err != nil {
		return err
	}
	if err := w.WriteField("name", emoji); err != nil {
		return err
	}
	if err := w.WriteField("mode", "data"); err != nil {
		return err
	}

	ext := "png"
	if contentType == "image/gif" {
		ext = "gif"
	}

	part, err := w.CreateFormFile("image", emoji+"."+ext)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil { // raw bytes straight in — no re-encode
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiUrl, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Cookie", "d="+b.State.Config.DCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to upload emoji: %s", resp.Status)
	}

	return nil
}
