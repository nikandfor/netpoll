//go:build ignore

package netpoll

import (
	"net"
	"syscall"

	"github.com/nikandfor/errors"
)

type (
	EventLoop struct {
		p *Poll

		l int

		h Handler

		Connected func(c *Conn) error
		Handler   func(c *Conn) error

		stopc chan struct{}
	}

	Conn struct {
		loop *EventLoop

		fd int
	}
)

func NewEventLoop() (e *EventLoop, err error) {
	p, err := NewPoll()
	if err != nil {
		return nil, errors.Wrap(err, "poll")
	}

	e = &EventLoop{
		p: p,
	}

	return e, nil
}

func (e *EventLoop) Serve(l net.Listener) (err error) {
	fd, err := Fd(l)
	if err != nil {
		return errors.Wrap(err, "get fd")
	}

	err = e.p.Add(fd, Read)
	if err != nil {
		return errors.Wrap(err, "add to poll")
	}

	e.l = fd

	return e.run()
}

func (e *EventLoop) run() error {
	var n int
	evs := make([]Event, 16)

	for {
		select {
		case <-e.stopc:
			return nil
		default:
		}

		n, err = e.p.Wait(evs, -1)
		if err != nil {
			return errors.Wrap(err, "wait")
		}

		for i := 0; i < n; i++ {
			e.procEvent(evs[i])
		}
	}
}

func (e *EventLoop) procEvent(ev Event) (err error) {
	if ev.Fd == e.l {
		for {
			fd, addr, err := syscall.Accept(ev.Fd)
			if err != nil {
				return errors.Wrap(err, "accept")
			}

			err = e.p.Add(fd, Read)
			if err != nil {
				return errors.Wrap(err, "accept: watch")
			}
		}
	}

}
