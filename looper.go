package dysel

import (
	"errors"
	"reflect"
)

type callbackType func(chosen int, recv reflect.Value, payload interface{}, recvOK bool) (continue_ bool)

type Looper struct {
	cases           *Cases
	callbackMap     map[reflect.Type]callbackType
	defaultCallback callbackType
	callbacks       []callbackType
}

var (
	ErrBadHandlerSignature = errors.New(
		"handler callback should be func(chosen int, recv reflect.Value, payload T, recvOK bool) (continue_ bool)")
	ErrAlreadySet = errors.New("payload type callback already set")
)

func (l *Looper) RecvAndCaseHandler(ch, payload interface{}, callback callbackType) error {
	payloadType := reflect.TypeOf(payload)
	err := l.AddCaseHandler(payloadType, callback)
	if err != nil {
		return err
	}
	l.cases.Recv(ch, payload)
	l.callbacks = append(l.callbacks, callback)
	return nil
}

func (l *Looper) Recv(ch, payload interface{}) {
	l.cases.Recv(ch, payload)
	l.callbacks = append(l.callbacks, nil)
}

func (l *Looper) Send(ch, value, payload interface{}) {
	l.cases.Send(ch, value, payload)
	l.callbacks = append(l.callbacks, nil)
}

func (l *Looper) AddCaseHandler(payloadType reflect.Type, callback callbackType) error {
	if l.callbackMap == nil {
		l.callbackMap = map[reflect.Type]callbackType{}
	}
	_, ok := l.callbackMap[payloadType]
	if ok {
		return ErrAlreadySet
	}

	l.callbackMap[payloadType] = callback
	return nil
}

func (l *Looper) Remove(chosen int) {
	l.cases.Remove(chosen)
	l.callbacks[chosen], l.callbacks[len(l.callbacks)-1] = l.callbacks[len(l.callbacks)-1], l.callbacks[chosen]
	l.callbacks = l.callbacks[:len(l.callbacks)-1]
}

func (l *Looper) SendNext(chosen int, val interface{}) {
	l.cases.SendNext(chosen, val)
}

func (l *Looper) Step() (continue_ bool) {
	chosen, recv, payload, recvOK := l.cases.DoSelect()
	callback := l.callbacks[chosen]
	if callback == nil {
		payloadType := reflect.TypeOf(payload)
		callback, ok := l.callbackMap[payloadType]
		if ok {
			l.callbacks[chosen] = callback
			return callback(chosen, recv, payload, recvOK)
		} else {
			return l.defaultCallback(chosen, recv, payload, recvOK)
		}
	} else {
		return callback(chosen, recv, payload, recvOK)
	}

}

func (l *Looper) Loop() {
	for l.Step() {
	}
}

func NewLooper(defaultCallback callbackType) *Looper {
	return &Looper{defaultCallback: defaultCallback, cases: &Cases{}}
}
