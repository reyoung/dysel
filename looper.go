package dysel

import (
	"errors"
	reflect "github.com/goccy/go-reflect"
)

type defaultCallbackType func(chosen int, recv reflect.Value, payload interface{}, recvOK bool) (continue_ bool)

type Looper struct {
	Cases           *Cases
	callbacks       map[reflect.Type]interface{}
	defaultCallback defaultCallbackType
}

var (
	ErrBadHandlerSignature = errors.New(
		"handler callback should be func(chosen int, recv reflect.Value, payload T, recvOK bool) (continue_ bool)")
	ErrAlreadySet = errors.New("payload type callback already set")
)

func (l *Looper) RecvAndCaseHandler(ch, payload, callback interface{}) error {
	payloadType := reflect.TypeOf(payload)
	err := l.AddCaseHandler(payloadType, callback)
	if err != nil {
		return err
	}
	l.Cases.Recv(ch, payload)
	return nil
}

func (l *Looper) AddCaseHandler(payloadType reflect.Type, callback interface{}) error {
	callbackType := reflect.TypeOf(callback)

	if callbackType.Kind() != reflect.Func || callbackType.NumOut() != 1 || callbackType.Out(
		0).Kind() != reflect.Bool || callbackType.NumIn() != 4 || callbackType.In(
		0).Kind() != reflect.Int || callbackType.In(1) != reflect.TypeOf(reflect.Value{}) || callbackType.In(
		2) != payloadType || callbackType.In(3).Kind() != reflect.Bool {
		return ErrBadHandlerSignature
	}
	if l.callbacks == nil {
		l.callbacks = map[reflect.Type]interface{}{}
	}
	_, ok := l.callbacks[payloadType]
	if ok {
		return ErrAlreadySet
	}

	l.callbacks[payloadType] = callback
	return nil
}

func (l *Looper) Step() (continue_ bool) {
	chosen, recv, payload, recvOK := l.Cases.DoSelect()
	payloadType := reflect.TypeOf(payload)
	callback, ok := l.callbacks[payloadType]
	if ok {
		results := reflect.ValueOf(callback).Call([]reflect.Value{reflect.ValueOf(chosen), reflect.ValueOf(recv),
			reflect.ValueOf(payload), reflect.ValueOf(recvOK)})
		return results[0].Bool()
	} else {
		return l.defaultCallback(chosen, recv, payload, recvOK)
	}
}

func (l *Looper) Loop() {
	for l.Step() {
	}
}

func NewLooper(defaultCallback defaultCallbackType) *Looper {
	return &Looper{defaultCallback: defaultCallback, Cases: &Cases{}}
}
