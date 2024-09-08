package internal

type AsyncDispatcher struct {
	messages   chan Message
	dispatcher Dispatcher
}

//func (d AsyncDispatcher) GetValue() int {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (d AsyncDispatcher) Produce(ctx context.Context, key string, payload []byte) error {
//	//TODO implement me
//	panic("implement me")
//}

type Message struct {
	Id int
}
type Dispatcher interface {
	IncrementCounter(idValue int)
}

func NewAsyncDispatcher(noOfWorkers int, bufSize int, dispatcher Dispatcher) AsyncDispatcher {
	d := AsyncDispatcher{messages: make(chan Message, bufSize), dispatcher: dispatcher}
	for i := 0; i < noOfWorkers; i++ {
		go d.asyncProcess()
	}
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
