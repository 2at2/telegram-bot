package telegram

type Handler interface {
	OnMessage(Pipe) error
	OnCallback(Pipe) error
	OnQuery(Pipe) error
	Test(Pipe) bool
}
