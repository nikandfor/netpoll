//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package netpoll

import (
	"syscall"
	"time"
)

func create() (int, error) {
	return syscall.Kqueue()
}

func eclose(fd int) error {
	return syscall.Close(fd)
}

func add(efd, fd, flags int) (err error) {
	ev := syscall.Kevent_t{
		Ident: uint64(fd),
		Flags: syscall.EV_ADD,
	}

	if flags&Read != 0 {
		ev.Filter |= syscall.EVFILT_READ
	}

	if flags&Write != 0 {
		ev.Filter |= syscall.EVFILT_WRITE
	}

	if ev.Filter == 0 {
		return ErrNoOperation
	}

	if flags&Oneshot != 0 {
		ev.Flags |= syscall.EV_ONESHOT
	}

	_, err = syscall.Kevent(efd, []syscall.Kevent_t{ev}, nil, nil)

	return
}

func del(efd, fd, flags int) (err error) {
	ev := syscall.Kevent_t{
		Ident: uint64(fd),
		Flags: syscall.EV_DELETE,
	}

	_, err = syscall.Kevent(efd, []syscall.Kevent_t{ev}, nil, nil)

	return
}

func wait(efd int, evs []Event, to time.Duration) (n int, err error) {
	raw := make([]syscall.Kevent_t, len(evs))

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

	n, err = syscall.Kevent(efd, nil, raw, t)
	if err = waitErr(err); err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		evs[i] = Event{
			Fd:   int(raw[i].Ident),
			Arg1: raw[i].Data,
		}

		if raw[i].Filter|syscall.EVFILT_READ != 0 {
			evs[i].Event |= Read
		}

		if raw[i].Filter|syscall.EVFILT_WRITE != 0 {
			evs[i].Event |= Write
		}
	}

	return
}
