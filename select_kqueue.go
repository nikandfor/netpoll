//go:build darwin || dragonfly || freebsd || netbsd || openbsd

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

	l := len(rs)
	ch := make([]syscall.Kevent_t, l+1)

	evs := ch[l:]
	ch = ch[:l]

	for i, r := range rs {
		fd, err := Fd(r)
		if err != nil {
			return nil, errors.Wrap(err, "fd %d", i)
		}

		ch[i] = syscall.Kevent_t{
			Ident:  uint64(fd),
			Filter: syscall.EVFILT_READ,
			Flags:  syscall.EV_ADD,
		}
	}

	var t *syscall.Timespec

	switch {
	case to < 0:
		t = nil // blocks indefinitely
	case to == 0:
		t = new(syscall.Timespec) // not blocks
	default:
		x := syscall.NsecToTimespec(int64(to))
		t = &x
	}

	//	tlog.Printw("kevent", "ch", ch, "timeout", to)

	n, err := syscall.Kevent(efd, ch, evs, t)
	if err = waitErr(err); err != nil {
		return nil, err
	}

	//	tlog.Printw("kevent", "evs", evs[:n])

	if n == 0 {
		return nil, ErrNoEvents
	}

	e := evs[0]

	if e.Flags&syscall.EV_ERROR != 0 {
		return nil, syscall.Errno(e.Data)
	}

	if e.Filter != syscall.EVFILT_READ {
		return nil, errors.New("unexpected filter: %v", e)
	}

	for j := 0; j < l; j++ {
		if e.Ident == ch[j].Ident {
			return rs[j], nil
		}
	}

	return nil, errors.New("unexpected fd")
}
