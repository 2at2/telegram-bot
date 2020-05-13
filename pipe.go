package telegram

import (
	"errors"
	api "github.com/2at2/telegram-api"
	"strings"
)

// Pipe main data holder
type Pipe interface {
	GetSender() *api.User
	GetChat() *api.Chat
	GetCallback() *api.Callback
	GetMessageId() int
	GetMessageText() string
	GetMessage() string
	GetCommand() string
	GetWhom() string
	GetCommandAndWhom() (string, string)
	SendMessage(message string, options *api.SendOptions) error
	EditMessageText(message string, options *api.SendOptions) error
	SendPhoto(photo *api.Photo, options *api.SendOptions) error
	SendCallbackAnswer(text string, alert bool) error
	DeleteMessage(messageId int) error
	SendTyping() error
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

// GetMessageText returns raw text
func (p *pipeHolder) GetMessageText() string {
	return p.message.Text
}

// GetMessage returns text without command
func (p *pipeHolder) GetMessage() string {
	text := p.GetMessageText()

	if strings.HasPrefix(text, "/") {
		parts := strings.Fields(text)
		text = strings.Join(parts[1:], " ")
	}

	return strings.TrimSpace(text)
}

// GetCommand returns command from text
func (p *pipeHolder) GetCommand() string {
	cmd, _ := p.getCommand()
	return cmd
}

// GetWhom returns name who should read the message
func (p *pipeHolder) GetWhom() string {
	_, whom := p.getCommand()
	return whom
}

// GetCommandAndWhom returns command from text
func (p *pipeHolder) GetCommandAndWhom() (string, string) {
	cmd, whom := p.getCommand()
	return cmd, whom
}

// GetCommand returns command from text
func (p *pipeHolder) getCommand() (string, string) {
	if p.message == nil || !strings.HasPrefix(p.message.Text, "/") {
		return "", ""
	}

	parts := strings.Fields(p.message.Text)

	whom := ""
	cmd := strings.TrimSpace(parts[0])

	if strings.Contains(cmd, "@") {
		x := strings.Split(cmd, "@")
		cmd = strings.TrimSpace(x[0])
		whom = strings.TrimSpace(x[1])
	}

	return cmd, whom
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
