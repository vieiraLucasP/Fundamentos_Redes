package queue

import (
	"fmt"
	"sync"

	"ring-network/pkg/message"
)

// MessageQueue implementa uma fila de mensagens thread-safe com tamanho máximo
type MessageQueue struct {
	messages []*message.QueuedMessage // Slice de mensagens na fila
	mutex    sync.RWMutex             // Mutex para acesso concorrente
	maxSize  int                      // Tamanho máximo da fila
}

// NewMessageQueue cria uma nova fila de mensagens com o tamanho máximo especificado
func NewMessageQueue(maxSize int) *MessageQueue {
	return &MessageQueue{
		messages: make([]*message.QueuedMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Enqueue adiciona uma nova mensagem à fila
// Retorna erro se a fila estiver cheia
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

// Dequeue remove e retorna a primeira mensagem da fila
// Retorna nil se a fila estiver vazia
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

// Peek retorna a primeira mensagem da fila sem removê-la
// Retorna nil se a fila estiver vazia
func (mq *MessageQueue) Peek() *message.QueuedMessage {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	if len(mq.messages) == 0 {
		return nil
	}

	return mq.messages[0]
}

// Size retorna o número atual de mensagens na fila
func (mq *MessageQueue) Size() int {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	return len(mq.messages)
}

// IsEmpty verifica se a fila está vazia
func (mq *MessageQueue) IsEmpty() bool {
	return mq.Size() == 0
}

// IsFull verifica se a fila está cheia
func (mq *MessageQueue) IsFull() bool {
	return mq.Size() >= mq.maxSize
}

// GetAll retorna uma cópia de todas as mensagens na fila
// Útil para exibir o estado atual da fila sem modificá-la
func (mq *MessageQueue) GetAll() []*message.QueuedMessage {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	result := make([]*message.QueuedMessage, len(mq.messages))
	copy(result, mq.messages)

	return result
}

// Clear remove todas as mensagens da fila
func (mq *MessageQueue) Clear() {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	mq.messages = mq.messages[:0]
}

// IncrementRetries incrementa o contador de tentativas da primeira mensagem
// Usado quando uma mensagem precisa ser retransmitida após um NAK
func (mq *MessageQueue) IncrementRetries() {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if len(mq.messages) > 0 {
		mq.messages[0].Retries++
	}
}

// GetFirstMessageRetries retorna o número de tentativas da primeira mensagem
func (mq *MessageQueue) GetFirstMessageRetries() int {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	if len(mq.messages) == 0 {
		return 0
	}

	return mq.messages[0].Retries
}

// RemoveFirstMessage remove e retorna a primeira mensagem da fila
// Alias para Dequeue para maior clareza semântica
func (mq *MessageQueue) RemoveFirstMessage() *message.QueuedMessage {
	return mq.Dequeue()
}

// String retorna uma representação em string da fila de mensagens
func (mq *MessageQueue) String() string {
	mq.mutex.RLock()
	defer mq.mutex.RUnlock()

	return fmt.Sprintf("MessageQueue{Size: %d/%d, Messages: %v}",
		len(mq.messages), mq.maxSize, mq.messages)
}
