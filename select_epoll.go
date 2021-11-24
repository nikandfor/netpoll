//go:build linux

package netpoll

import (
	"io"
	"syscall"
	"time"

	"github.com/nikandfor/errors"
)

func Select(rs []io.Reader, to time.Duration) (io.Reader, error) {
	efd, err := create()
	if err != nil {
		return nil, errors.Wrap(err, "new poll")
	}

	defer func() {
		e := eclose(efd)
		if err == nil {
			err = errors.Wrap(e, "close poll")
		}
	}()

	fds := make([]int, len(rs))

	for i, r := range rs {
		fd, err := Fd(r)
		if err != nil {
			return nil, errors.Wrap(err, "fd %d", i)
		}

		fds[i] = fd

		err = add(efd, fd, Read)
		if err != nil {
			return nil, errors.Wrap(err, "add %d", i)
		}
	}

	msec := 0

	switch {
	case to < 0:
		msec = -1 // blocks indefinitely
	case to == 0:
		msec = 0 // not blocks
	default:
		msec = int(to / time.Millisecond)
	}

	var evs [1]syscall.EpollEvent

	n, err := syscall.EpollWait(efd, evs[:], msec)
	if err = waitErr(err); err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNoEvents
	}

	e := evs[0]

	if e.Events == syscall.EPOLLERR {
		return nil, errors.New("file error")
	}

	switch {
	case e.Events == syscall.EPOLLIN:
		// ready to read
	case e.Events&syscall.EPOLLHUP != 0 || e.Events&syscall.EPOLLRDHUP != 0:
		// reader closed or reader+writer closed
	default:
		return nil, errors.New("unexpected filter: %v", e)
	}

	for j := 0; j < len(rs); j++ {
		if int(e.Fd) == fds[j] {
			return rs[j], nil
		}
	}

	return nil, errors.New("unexpected fd")
}
