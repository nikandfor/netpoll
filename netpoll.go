package netpoll

import (
	"syscall"
	"time"

	"github.com/nikandfor/errors"
)

type (
	Poll struct {
		fd int
	}

	Event struct {
		Fd    int
		Event int
		Arg1  int64
	}
)

const (
	_ = 1 << iota

	EOF
	Read
	Write
	Closed

	Oneshot
)

var (
	ErrNoOperation = errors.New("no operation selected")
	ErrUnsupported = errors.New("unsupported")
	ErrNoEvents    = errors.New("no events")
)

func NewPoll() (p *Poll, err error) {
	fd, err := create()
	if err != nil {
		return nil, errors.Wrap(err, "create")
	}

	p = &Poll{
		fd: fd,
	}

	return p, nil
}

func (p *Poll) Close() (err error) {
	return eclose(p.fd)
}

func (p *Poll) Add(fd, flags int) (err error) {
	return add(p.fd, fd, flags)
}

func (p *Poll) Del(fd int) (err error) {
	return del(p.fd, fd, 0)
}

// Wait waits for events storing them to evs buffer.
// returned n is a number of events read the same as io.Reader works.
// to is timeout to wait.
//
// If to is -1 it blocks until some event happen or interrupted.
// If to is 0 it returns immidietly even if no events occured.
// Otherwise it's a maximum time to wait.
func (p *Poll) Wait(evs []Event, to time.Duration) (n int, err error) {
	return wait(p.fd, evs, to)
}

func Fd(f interface{}) (fd int, err error) {
	if fl, ok := f.(interface {
		Fd() uintptr
	}); ok {
		return int(fl.Fd()), nil
	}

	if fl, ok := f.(interface {
		Fd() int
	}); ok {
		return fl.Fd(), nil
	}

	if sc, ok := f.(interface {
		SyscallConn() (syscall.RawConn, error)
	}); ok {
		rc, err := sc.SyscallConn()
		if err != nil {
			return -1, err
		}

		err = rc.Control(func(f uintptr) {
			fd = int(f)
		})

		return fd, err
	}

	return -1, ErrUnsupported
}

func waitErr(err error) error {
	switch err {
	case nil:
		return nil
	case syscall.EINTR:
		return ErrNoEvents
	default:
		return errors.Wrap(err, "wait")
	}
}

func noEINTR(err error) error {
	switch err {
	case syscall.EINTR:
		return ErrNoEvents
	default:
		return err
	}
}
