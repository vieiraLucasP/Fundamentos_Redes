package queue

import (
	"fmt"
	"sync"

	"ring-network/pkg/message"
)

type MessageQueue struct {
	messages []*message.QueuedMessage
	mutex    sync.RWMutex
	maxSize  int
}

func NewMessageQueue(maxSize int) *MessageQueue {
	return &MessageQueue{
		messages: make([]*message.QueuedMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

func (mq *MessageQueue) Enqueue(destination, content string) error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.messages) >= mq.maxSize {
		return fmt.Errorf("fila cheia (máximo: %d mensagens)", mq.maxSize)
	}

	queuedMsg := message.NewQueuedMessage(destination, content)
	mq.messages = append(mq.messages, queuedMsg)

	return nil
}

func (mq *MessageQueue) Dequeue() *message.QueuedMessage {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.messages) == 0 {
		return nil
	}

	msg := mq.messages[0]
	mq.messages = mq.messages[1:]

	return msg
}

func (mq *MessageQueue) Peek() *message.QueuedMessage {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	if len(mq.messages) == 0 {
		return nil
	}

	return mq.messages[0]
}

func (mq *MessageQueue) Size() int {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	return len(mq.messages)
}

func (mq *MessageQueue) IsEmpty() bool {
	return mq.Size() == 0
}

func (mq *MessageQueue) IsFull() bool {
	return mq.Size() >= mq.maxSize
}

func (mq *MessageQueue) GetAll() []*message.QueuedMessage {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	result := make([]*message.QueuedMessage, len(mq.messages))
	copy(result, mq.messages)

	return result
}

func (mq *MessageQueue) Clear() {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	mq.messages = mq.messages[:0]
}

func (mq *MessageQueue) IncrementRetries() {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.messages) > 0 {
		mq.messages[0].Retries++
	}
}

func (mq *MessageQueue) GetFirstMessageRetries() int {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	if len(mq.messages) == 0 {
		return 0
	}

	return mq.messages[0].Retries
}

func (mq *MessageQueue) RemoveFirstMessage() *message.QueuedMessage {
	return mq.Dequeue()
}

func (mq *MessageQueue) String() string {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	return fmt.Sprintf("MessageQueue{Size: %d/%d, Messages: %v}",
		len(mq.messages), mq.maxSize, mq.messages)
}
