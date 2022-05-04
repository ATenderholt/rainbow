package domain

type SqsQueue struct {
	ID         int
	Name       string
	Tags       []SqsQueueTag       `ref:"id" fk:"queue_id" auto:"true"`
	Attributes []SqsQueueAttribute `ref:"id" fk:"queue_id" auto:"true"`
}

func NewSqsQueue(name string) SqsQueue {
	return SqsQueue{
		Name:       name,
		Tags:       nil,
		Attributes: nil,
	}
}

func (queue *SqsQueue) AddAttribute(name string, value string) {
	attr := SqsQueueAttribute{
		Name:  name,
		Value: value,
	}

	queue.Attributes = append(queue.Attributes, attr)
}

func (queue *SqsQueue) AddTag(name string, value string) {
	tag := SqsQueueTag{
		Name:  name,
		Value: value,
	}

	queue.Tags = append(queue.Tags, tag)
}

type SqsQueueAttribute struct {
	ID      int
	QueueId int
	Name    string
	Value   string
}

type SqsQueueTag struct {
	ID      int
	QueueId int
	Name    string
	Value   string
}
