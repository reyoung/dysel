package dysel

import (
	"context"
	"sync"
	"testing"
	"time"
	
	reflect "github.com/goccy/go-reflect"
)

func TestPingPongUntilContextDone(t *testing.T) {
	pingChan := make(chan int, 1)
	pongChan := make(chan int, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	type ctxDonePayload struct{}
	type pingPayload struct{}
	type pongPayload struct{}
	type sentPayload struct{}

	looper := NewLooper(func(chosen int, recv reflect.Value, payload interface{}, recvOK bool) (continue_ bool) {
		t.FailNow()
		return false
	})

	err := looper.AddCaseHandler(reflect.TypeOf(sentPayload{}),
		func(chosen int, _ reflect.Value, _ sentPayload, _ bool) bool {
			looper.Cases.Remove(chosen)
			return true
		})

	if err != nil {
		t.FailNow()
	}

	err = looper.RecvAndCaseHandler(ctx.Done(), ctxDonePayload{}, func(int, reflect.Value, ctxDonePayload, bool) bool {
		return false
	})
	if err != nil {
		t.FailNow()
	}

	err = looper.RecvAndCaseHandler(pingChan, pingPayload{},
		func(_ int, val reflect.Value, _ pingPayload, _ bool) bool {
			recv := val.Interface().(int)
			if recv%2 == 0 {
				t.FailNow()
			}
			looper.Cases.Send(pongChan, recv+1, sentPayload{})
			return true
		})

	err = looper.RecvAndCaseHandler(pongChan, pongPayload{},
		func(_ int, val reflect.Value, _ pongPayload, _ bool) bool {
			recv := val.Interface().(int)
			looper.Cases.Send(pingChan, recv+1, sentPayload{})
			return true
		})

	var completeWG sync.WaitGroup
	completeWG.Add(1)
	go func() {
		defer completeWG.Done()
		looper.Loop()
	}()
	pingChan <- 1
	completeWG.Wait()
}
