package dysel

import reflect "github.com/goccy/go-reflect"

// Cases maintain a list of select cases for dynamic selection.
//
// Each case is associated with an interface as a payload. `DoSelect` returns the associated payload as well.
//
// The underlying cases is **UNORDERED**. You can use `payload` to figure out which case is selected.
type Cases struct {
	cases    []reflect.SelectCase
	payloads []interface{}
}

func (c *Cases) Recv(ch interface{}, payload interface{}) {
	c.cases = append(c.cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ch),
	})
	c.payloads = append(c.payloads, payload)
}

func (c *Cases) Send(ch, val, payload interface{}) {
	c.cases = append(c.cases, reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: reflect.ValueOf(ch),
		Send: reflect.ValueOf(val),
	})
	c.payloads = append(c.payloads, payload)
}

func (c *Cases) Remove(chosen int) {
	c.cases[chosen], c.cases[len(c.cases)-1] = c.cases[len(c.cases)-1], c.cases[chosen]
	c.payloads[chosen], c.payloads[len(c.payloads)-1] = c.payloads[len(c.payloads)-1], c.payloads[chosen]
	c.cases = c.cases[:len(c.cases)-1]
	c.payloads = c.payloads[:len(c.payloads)-1]
}

// SendNext replaces the sending value.
func (c *Cases) SendNext(chosen int, val interface{}) {
	if c.cases[chosen].Dir != reflect.SelectSend {
		panic("should only call send next in send callback.")
	}
	c.cases[chosen].Send = reflect.ValueOf(val)
}

func (c *Cases) DoSelect() (chosen int, recv reflect.Value, payload interface{}, recvOK bool) {
	chosen, recv, recvOK = reflect.Select(c.cases)
	payload = c.payloads[chosen]
	return
}
