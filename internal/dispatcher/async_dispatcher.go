package dispatcher

type AsyncDispatcher struct {
	messages   chan Message
	dispatcher Dispatcher
}

type Message struct {
	Id int
}
type Dispatcher interface {
	GetValue() int
	IncrementCounter(idValue int)
}

func NewAsyncDispatcher(noOfWorkers int, bufSize int, dispatcher Dispatcher) AsyncDispatcher {
	d := AsyncDispatcher{messages: make(chan Message, bufSize), dispatcher: dispatcher}
	for i := 0; i < noOfWorkers; i++ {
		go d.asyncProcess()
	}

	//go d.monitorChannel()

	return d
}

func (d AsyncDispatcher) Dispatch(m Message) {
	d.messages <- m
}

func (d AsyncDispatcher) asyncProcess() {
	for msg := range d.messages {
		d.dispatcher.IncrementCounter(msg.Id)
	}
}
