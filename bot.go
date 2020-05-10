package telegram

import (
	"errors"
	api "github.com/2at2/telegram-api"
	"github.com/sirupsen/logrus"
	"sync"
)

type bot struct {
	*logrus.Logger

	bot      *api.Bot
	handlers []Handler
	m        sync.Mutex

	postIncomingMessageListener  []func(pipe Pipe)
	postIncomingCallbackListener []func(pipe Pipe)
}

// New returns new bot
func New(
	b *api.Bot,
	logger *logrus.Logger,
) (*bot, error) {
	if b == nil {
		return nil, errors.New("nil bot given")
	}

	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &bot{
		Logger: logger,
		bot:    b,
	}, nil
}

// AddHandler adds handlers
func (b *bot) AddHandler(h Handler) {
	b.handlers = append(b.handlers, h)
}

// AddPostIncomingMessageListener adds listener
func (b *bot) AddPostIncomingMessageListener(f ...func(pipe Pipe)) {
	b.postIncomingMessageListener = append(b.postIncomingMessageListener, f...)
}

// AddPostIncomingCallbackListener adds listener
func (b *bot) AddPostIncomingCallbackListener(f ...func(pipe Pipe)) {
	b.postIncomingCallbackListener = append(b.postIncomingCallbackListener, f...)
}

// OnIncomingMessage handle message
func (b *bot) OnIncomingMessage(message api.Message) {
	b.Debugf("Received message %d from %d", message.ID, message.Sender.ID)

	pipe, err := NewPipe(
		&message,
		nil,
		b.bot,
	)

	if err != nil {
		b.Errorf("Unable to build pipe - %s", err)
		return
	}

	_ = pipe.SendTyping()

	if err := b.route(pipe); err != nil {
		b.Errorf("Unable to route message - %s", err)
	}

	b.Debugf("Message %d is processed", pipe.GetMessageId())

	// Invoking listeners
	for _, x := range b.postIncomingMessageListener {
		x(pipe)
	}
}

// OnIncomingCallback handle callback
func (b *bot) OnIncomingCallback(callback api.Callback) {
	b.Debugf("Received callback %s from %d", callback.ID, callback.Sender.ID)

	pipe, err := NewPipe(
		nil,
		&callback,
		b.bot,
	)

	if err != nil {
		b.Errorf("Unable to build pipe - %s", err)
		return
	}

	if err := b.route(pipe); err != nil {
		b.Errorf("Unable to route callback - %s", err)
	}

	b.Debugf("Callback %d is processed", pipe.GetMessageId())

	// Invoking listeners
	for _, x := range b.postIncomingMessageListener {
		x(pipe)
	}
}

// route routes request
func (b *bot) route(pipe Pipe) error {
	b.Trace("Route is invoked")

	handler := b.findHandler(pipe)

	if handler == nil {
		b.Info("Undefined route") // TODO
	} else {
		b.Trace("Handler is found")

		var err error
		if pipe.GetCallback() != nil {
			b.Trace("On callback event")
			err = handler.OnCallback(pipe)
		} else {
			b.Trace("On message event")
			err = handler.OnMessage(pipe)
		}

		if err != nil {
			b.Errorf("Error occurred - %s", err)
			_ = pipe.SendMessage("Error - "+err.Error(), nil)
		}
	}

	return nil
}

// findHandler returns route
func (b *bot) findHandler(pipe Pipe) Handler {
	for _, handler := range b.handlers {
		if handler.Test(pipe) {
			return handler
		}
	}
	return nil
}
