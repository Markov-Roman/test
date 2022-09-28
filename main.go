package main

import (
	"fmt"
	"sync"
	"time"
)

const SizeBuf int = 6
const Delay int = 5

type RingBuf struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

func NewRingBuf(size int) *RingBuf {
	return &RingBuf{array: make([]int, size), pos: -1, size: size, m: sync.Mutex{}}
}

func (r *RingBuf) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = el
		return
	}
	r.pos++
	r.array[r.pos] = el
}

func (r *RingBuf) Get() []int {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos <= 0 {
		return nil
	}
	pos := r.pos
	r.pos = -1
	return r.array[:pos+1]
}

func Init(done chan int) <-chan int {
	var elem int
	input := make(chan int)
	go func() {
		defer close(input)
		for {
			_, err := fmt.Scanln(&elem)
			if err != nil {
				continue
			}
			select {
			case input <- elem:
			case <-done:
				return
			}

		}
	}()
	return input
}

func FilterMinus(done <-chan int, input <-chan int) <-chan int {
	filteredStream := make(chan int)
	go func() {
		defer close(filteredStream)
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				if i >= 0 {
					select {
					case filteredStream <- i:
					case <-done:
						return
					}
				}
			}
		}
	}()
	return filteredStream
}

func FilterUnDiv3(done <-chan int, input <-chan int) <-chan int {
	filteredStream := make(chan int)
	go func() {
		defer close(filteredStream)
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				if (i != 0) && (i%3 == 0) {
					select {
					case filteredStream <- i:
					case <-done:
						return
					}
				}
			}
		}
	}()
	return filteredStream
}

func Buffer(done <-chan int, input <-chan int) <-chan int {
	buf := NewRingBuf(SizeBuf)
	nextChan := make(chan int)
	go func() {
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				buf.Push(i)
			}
		}
	}()
	go func() {
		defer close(nextChan)
		for {
			time.Sleep(time.Second * time.Duration(Delay))
			out := buf.Get()
			for _, val := range out {
				select {
				case nextChan <- val:
				case <-done:
					return
				}
			}
		}
	}()
	return nextChan
}

func main() {
	done := make(chan int)
	defer close(done)
	intStream := Init(done)
	pipeline := Buffer(done, FilterUnDiv3(done, FilterMinus(done, intStream)))
	for v := range pipeline {
		fmt.Println("Получено", v)
	}
}
