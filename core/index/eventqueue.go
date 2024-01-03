package index

type EventQueue struct {
	closed bool
	queue  chan Event
	writer *Writer
}

func NewEventQueue(writer *Writer) *EventQueue {
	return &EventQueue{
		closed: false,
		queue:  make(chan Event, 10),
		writer: writer,
	}
}

type Event func(writer *Writer) error

func (e *EventQueue) Add(event Event) bool {
	select {
	case e.queue <- event:
		return true
	default:
		return false
	}
}

func (e *EventQueue) processEvents() error {
OUT:
	for {
		select {
		case fn := <-e.queue:
			if err := fn(e.writer); err != nil {
				return err
			}
		default:
			break OUT
		}
	}
	return nil
}
