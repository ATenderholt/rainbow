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
	ID      int `xml:"-"`
	QueueId int `xml:"-"`
	Name    string
	Value   string
}

type SqsQueueTag struct {
	ID      int `xml:"-"`
	QueueId int `xml:"-"`
	Name    string
	Value   string
}

type GetQueueAttributesResult struct {
	Attribute []SqsQueueAttribute
}

type ResponseMetadata struct {
	RequestId string
}

type GetQueueAttributesResponse struct {
	GetQueueAttributesResult GetQueueAttributesResult
	ResponseMetadata         ResponseMetadata
}

func (obj *GetQueueAttributesResponse) AddAttributeIfNotExists(key, value string) {
	for _, attribute := range obj.GetQueueAttributesResult.Attribute {
		if attribute.Name == key {
			return
		}
	}

	obj.GetQueueAttributesResult.Attribute = append(obj.GetQueueAttributesResult.Attribute,
		SqsQueueAttribute{Name: key, Value: value},
	)
}
