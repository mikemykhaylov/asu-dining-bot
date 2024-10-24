package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

var (
	TelegramBaseURL          string = "https://api.telegram.org/bot"
	TelegramPublicPersonalID int64  = -1
)

type TelegramAPI struct {
	baseURL    string
	httpClient *http.Client
}

func NewTelegramAPI(token string, httpClient *http.Client) *TelegramAPI {
	return &TelegramAPI{
		baseURL:    TelegramBaseURL + token,
		httpClient: httpClient,
	}
}

func (t *TelegramAPI) SendMessage(ctx context.Context, chatID int64, text string) error {
	requestURL, err := url.Parse(fmt.Sprintf("%s/sendMessage", t.baseURL))
	if err != nil {
		return err
	}
	query := requestURL.Query()
	query.Set("chat_id", fmt.Sprintf("%d", chatID))
	query.Set("text", text)
	query.Set("parse_mode", "HTML")
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return err
	}

	res, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (t *TelegramAPI) SendVideo(ctx context.Context, chatID int64, videoPath string) error {
	requestURL, _ := url.Parse(fmt.Sprintf("%s/sendVideo", t.baseURL))
	query := requestURL.Query()
	query.Set("chat_id", fmt.Sprintf("%d", chatID))
	requestURL.RawQuery = query.Encode()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("video", videoPath)
	if err != nil {
		return err
	}

	file, err := os.Open(videoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, requestURL.String(), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}
