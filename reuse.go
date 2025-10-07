package main

import (
	"fmt"

	"capnproto.org/go/capnp/v3"
)

type CapnpReuse struct {
	readMsg *capnp.Message
	readBuf []byte
}

func NewCReuse() *CapnpReuse {
	msg, _ := capnp.NewSingleSegmentMessage(nil)
	return &CapnpReuse{
		readMsg: msg,
		readBuf: make([]byte, 8*1024),
	}
}

func (s *CapnpReuse) ReadMessage(data []byte) error {
	msg, err := capnp.Unmarshal(data)
	if err != nil {
		return err
	}

	s.readMsg = msg
	_, err = s.readMsg.Segment(0)
	if err != nil {
		return err
	}
	return nil
}

func (s *CapnpReuse) ReadMessageZero(data []byte) error {
	err := capnp.UnmarshalZeroThree(s.readMsg, data)
	if err != nil {
		return err
	}
	_, err = s.readMsg.Segment(0)
	if err != nil {
		return err
	}
	return nil
}

func (s *CapnpReuse) ReadMessageZeroTo(data []byte) error {
	err := capnp.UnmarshalZeroTo(s.readMsg, &s.readBuf, data)
	if err != nil {
		return err
	}
	_, err = s.readMsg.Segment(0)
	if err != nil {
		return err
	}
	return nil
}

func Deserialize[T any](s *CapnpReuse, data []byte, get func(*capnp.Message) (T, error)) (T, error) {
	err := s.ReadMessage(data)

	if err != nil {
		var zero T
		return zero, fmt.Errorf("what: %w", err)
	}
	return get(s.readMsg)
}

func DeserializeZero[T any](s *CapnpReuse, data []byte, get func(*capnp.Message) (T, error)) (T, error) {
	err := s.ReadMessageZero(data)
	if err != nil {
		var zero T
		return zero, err
	}
	return get(s.readMsg)
}

func DeserializeZeroTo[T any](s *CapnpReuse, data []byte, get func(*capnp.Message) (T, error)) (T, error) {
	err := s.ReadMessageZeroTo(data)
	if err != nil {
		var zero T
		return zero, err
	}
	return get(s.readMsg)
}
