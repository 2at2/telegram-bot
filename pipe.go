package telegram

import (
	"errors"
	api "github.com/2at2/telegram-api"
	"strings"
)

// Pipe main data holder
type Pipe interface {
	GetCallback() *api.Callback
	GetMessageId() int
	GetMessage() string
	GetCommand() string
	SendMessage(message string, options *api.SendOptions) error
	EditMessageText(message string, options *api.SendOptions) error
	SendPhoto(photo *api.Photo, options *api.SendOptions) error
	SendCallbackAnswer(text string, alert bool) error
	DeleteMessage(messageId int) error
}

// pipeHolder pipe implementation
type pipeHolder struct {
	message  *api.Message
	callback *api.Callback

	bot    *api.Bot
	chat   api.Chat
	sender api.User
}

// NewPipe returns pipe
func NewPipe(
	message *api.Message,
	callback *api.Callback,
	bot *api.Bot,
) (*pipeHolder, error) {
	if message == nil && callback == nil {
		return nil, errors.New("expected message or callback nil given")
	}
	if bot == nil {
		return nil, errors.New("empty bot api given")
	}

	if message == nil {
		message = &callback.Message
	}

	return &pipeHolder{
		message:  message,
		callback: callback,
		bot:      bot,
		chat:     message.Chat,
		sender:   message.Sender,
	}, nil
}

// GetMessageId returns id of message
func (p *pipeHolder) GetMessageId() int {
	return p.message.ID
}

// GetMessage returns raw text
func (p *pipeHolder) GetMessage() string {
	return p.message.Text
}

// GetCommand returns command from text
func (p *pipeHolder) GetCommand() string {
	if p.message == nil {
		return ""
	}

	parts := strings.Fields(p.message.Text)
	return "/" + strings.TrimLeft(strings.TrimSpace(parts[0]), "/")
}

// GetCallback returns callback
func (p *pipeHolder) GetCallback() *api.Callback {
	return p.callback
}

// GetSender returns sender
func (p *pipeHolder) GetSender() *api.User {
	if p.message != nil {
		return &p.message.Sender
	} else if p.callback != nil {
		return &p.callback.Sender
	}

	return nil
}

// GetSender returns sender
func (p *pipeHolder) GetChat() *api.Chat {
	if p.message != nil {
		return &p.message.Chat
	} else if p.callback != nil {
		return &p.callback.Message.Chat
	}

	return nil
}

// SendMessage sends message to chat
func (p *pipeHolder) SendMessage(message string, options *api.SendOptions) error {
	return p.bot.SendMessage(p.chat, message, embedSendOptions(options))
}

// EditMessageText sends edited text
func (p *pipeHolder) EditMessageText(message string, options *api.SendOptions) error {
	return p.bot.EditMessageText(p.chat, message, embedSendOptions(options))
}

// SendPhoto sends photo to chat
func (p *pipeHolder) SendPhoto(photo *api.Photo, options *api.SendOptions) error {
	return p.bot.SendPhoto(p.chat, photo, embedSendOptions(options))
}

// SendCallbackAnswer sends callback answer
func (p *pipeHolder) SendCallbackAnswer(text string, alert bool) error {
	return p.bot.AnswerCallbackQuery(
		p.callback,
		&api.CallbackResponse{
			Text:      text,
			ShowAlert: alert,
		},
	)
}

// DeleteMessage removes message
func (p *pipeHolder) DeleteMessage(messageId int) error {
	return p.bot.DeleteMessage(p.chat, messageId)
}

// SendTyping sends action typing
func (p *pipeHolder) SendTyping() error {
	return p.bot.SendChatAction(p.chat, api.Typing)
}

// embedSendOptions patches options
func embedSendOptions(options *api.SendOptions) *api.SendOptions {
	if options == nil {
		options = &api.SendOptions{
			ParseMode: api.ModeDefault,
		}
	}

	if options.ParseMode == api.ModeDefault {
		options.ParseMode = api.ModeMarkdown
	}

	return options
}
