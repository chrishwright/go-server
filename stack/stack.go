package stack

import (
	"fmt"
	"sync"
)

type Stack struct {
	stack [][]byte
	mutex sync.Mutex
}

func (s *Stack) IsFull() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.stack) == 100
}

func (s *Stack) IsEmpty() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return len(s.stack) == 0
}

// Pushes an array of bytes to the stack.
// Note: this is not a thread-safe function.
func (s *Stack) PushToStack(bytesToAdd []byte) {
	s.stack = append(s.stack, bytesToAdd)
	fmt.Println("The size of the stack is now: ", len(s.stack))
}

// Returns an array of bytes from the stack.
// Note: this is not a thread-safe function.
func (s *Stack) PopFromStack() []byte {
	response := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	fmt.Println("The size of the stack is now: ", len(s.stack))
	return response
}
