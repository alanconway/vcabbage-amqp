// +build local

package amqp_test

import (
	"context"
	"net"
	"testing"

	"pack.ag/amqp"
)

// Tests that require a local broker running on the standard AMQP port.

func TestDial_IPV6(t *testing.T) {
	c, err := amqp.Dial("amqp://localhost")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()

	l, err := net.Listen("tcp6", "[::]:0")
	if err != nil {
		t.Skip("ipv6 not supported")
	}
	l.Close()

	for _, u := range []string{"amqp://[::]:5672", "amqp://[::]"} {
		u := u // Don't  use range variable in func literal.
		t.Run(u, func(t *testing.T) {
			c, err := amqp.Dial(u)
			if err != nil {
				t.Errorf("%q: %v", u, err)
			} else {
				c.Close()
			}
		})
	}
}

func TestSendReceive(t *testing.T) {
	c, err := amqp.Dial("amqp://")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	ssn, err := c.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	r, err := ssn.NewReceiver(amqp.LinkAddressDynamic())
	if err != nil {
		t.Fatal(err)
	}
	s, err := ssn.NewSender(amqp.LinkAddress(r.Address()))
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error)
	go func() {
		done <- s.Send(context.Background(), amqp.NewMessage([]byte("hello")))
	}()
	defer func() {
		err := <-done
		if err != nil {
			t.Error(err)
		}
	}()

	m, err := r.Receive(context.Background())
	if err != nil {
		t.Error(err)
	} else {
		m.Accept()
	}
	if "hello" != string(m.GetData()) {
		t.Errorf("\"hello\" != %#v", m.GetData())
	}
}
