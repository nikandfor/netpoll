//go:build linux

package netpoll

import (
	"syscall"
	"time"
)

func create() (int, error) {
	return syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
}

func eclose(fd int) error {
	return syscall.Close(fd)
}

func add(efd, fd, flags int) error {
	ev := &syscall.EpollEvent{
		Fd: int32(fd),
	}

	if flags&Read != 0 {
		ev.Events |= syscall.EPOLLIN
	}

	if flags&Write != 0 {
		ev.Events |= syscall.EPOLLOUT
	}

	if ev.Events == 0 {
		return ErrNoOperation
	}

	if flags&Oneshot != 0 {
		ev.Events |= syscall.EPOLLONESHOT
	}

	return syscall.EpollCtl(efd, syscall.EPOLL_CTL_ADD, fd, ev)
}

func del(efd, fd, flags int) error {
	ev := &syscall.EpollEvent{
		Fd: int32(fd),
	}

	return syscall.EpollCtl(efd, syscall.EPOLL_CTL_DEL, fd, ev)
}

func wait(efd int, evs []Event, to time.Duration) (n int, err error) {
	raw := make([]syscall.EpollEvent, len(evs))

	msec := 0

	switch {
	case to < 0:
		msec = -1 // blocks indefinitely
	case to == 0:
		msec = 0 // not blocks
	default:
		msec = int(to / time.Millisecond)
	}

	n, err = syscall.EpollWait(efd, raw, msec)
	if err = waitErr(err); err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		evs[i] = Event{
			Fd: int(raw[i].Fd),
		}

		if raw[i].Events|syscall.EPOLLIN != 0 {
			evs[i].Event |= Read
		}

		if raw[i].Events|syscall.EPOLLOUT != 0 {
			evs[i].Event |= Write
		}
	}

	return
}
