package telegram

import (
	"errors"
	api "github.com/2at2/telegram-api"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type bot struct {
	*logrus.Logger

	bot             *api.Bot
	botName         string
	filterByBotName bool
	handlers        []Handler
	m               sync.Mutex

	messages  chan api.Message
	queries   chan api.Query
	callbacks chan api.Callback

	preListener  []func(pipe Pipe) bool
	postListener []func(pipe Pipe)
}

// New returns new bot
func New(
	botName string,
	filterByBotName bool,
	b *api.Bot,
	logger *logrus.Logger,
) (*bot, error) {
	if b == nil {
		return nil, errors.New("nil bot given")
	}
	if len(botName) == 0 {
		return nil, errors.New("empty bot name")
	}
	if logger == nil {
		logger = logrus.StandardLogger()
	}

	return &bot{
		Logger:          logger,
		bot:             b,
		botName:         botName,
		filterByBotName: filterByBotName,
		messages:        make(chan api.Message, 1),
		queries:         make(chan api.Query, 1),
		callbacks:       make(chan api.Callback, 1),
	}, nil
}

// AddHandler adds handlers
func (b *bot) AddHandler(h Handler) {
	b.handlers = append(b.handlers, h)
}

// AddPostIncomingMessageListener adds listener
func (b *bot) AddPreListener(f ...func(pipe Pipe) bool) {
	b.preListener = append(b.preListener, f...)
}

// AddPostIncomingMessageListener adds listener
func (b *bot) AddPostListener(f ...func(pipe Pipe)) {
	b.postListener = append(b.postListener, f...)
}

// OnIncomingMessage handle message
func (b *bot) OnIncomingMessage(message api.Message) {
	b.Tracef("Received message %d from %d", message.ID, message.Sender.ID)

	pipe, err := NewPipe(
		&message,
		nil,
		b.bot,
	)

	if err != nil {
		b.Errorf("Unable to build pipe - %s", err)
		return
	}

	// Bot name filter
	if b.filterByBotName && pipe.GetWhom() != b.botName {
		b.Trace("Skip by bot name")
		return
	}

	// Invoking listeners
	for _, x := range b.preListener {
		if cont := x(pipe); !cont {
			return
		}
	}

	if err := b.route(pipe); err != nil {
		b.Errorf("Unable to route message - %s", err)
	}

	b.Tracef("Message %d is processed", pipe.GetMessageId())

	// Invoking listeners
	for _, x := range b.postListener {
		x(pipe)
	}
}

// OnIncomingCallback handle callback
func (b *bot) OnIncomingCallback(callback api.Callback) {
	b.Tracef("Received callback %s from %d", callback.ID, callback.Sender.ID)

	pipe, err := NewPipe(
		nil,
		&callback,
		b.bot,
	)

	if err != nil {
		b.Errorf("Unable to build pipe - %s", err)
		return
	}

	// Bot name filter
	if b.filterByBotName && pipe.GetWhom() != b.botName {
		b.Trace("Skip by bot name")
		return
	}

	// Invoking listeners
	for _, x := range b.preListener {
		if cont := x(pipe); !cont {
			return
		}
	}

	if err := b.route(pipe); err != nil {
		b.Errorf("Unable to route callback - %s", err)
	}

	b.Tracef("Callback %d is processed", pipe.GetMessageId())

	// Invoking listeners
	for _, x := range b.postListener {
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

		_ = pipe.SendTyping()

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

func (b *bot) Start(poll time.Duration, stop chan bool, wgg *sync.WaitGroup) {
	defer wgg.Done()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go b.bot.Listen(b.messages, b.queries, b.callbacks, poll, stop, wg)

	// Reading messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			case message := <-b.messages:
				b.Tracef("Received telegram message - %d", message.ID)
				go b.OnIncomingMessage(message)
			}
		}
	}()

	// Reading callbacks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			case callback := <-b.callbacks:
				b.Tracef("Received telegram callback - %s", callback.ID)
				go b.OnIncomingCallback(callback)
			}
		}
	}()

	// Reading callbacks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			case query := <-b.queries:
				b.Tracef("Received telegram query - %s", query.ID)
			}
		}
	}()

	wg.Wait()
}
