package store

import (
	"errors"
	"io"
)

const (
	MSG_LOCK_RELEASED = 0
	MSG_LOCK_ACQUIRED = 1
)

var _ LockFactory = &VerifyingLockFactory{}

// VerifyingLockFactory A LockFactory that wraps another LockFactory and verifies that each lock
// obtain/release is "correct" (never results in two processes holding the lock at the same time).
// It does this by contacting an external server (LockVerifyServer) to assert that at most one process
// holds the lock at a time. To use this, you should also run LockVerifyServer on the host and port
// matching what you pass to the constructor.
//
// See Also: 	LockVerifyServer,
//
//	LockStressTest
type VerifyingLockFactory struct {
	lf  LockFactory
	in  io.ByteReader
	out io.ByteWriter
}

// NewVerifyingLockFactory Params:
// * lf – the LockFactory that we are testing
// * in – the socket's input to LockVerifyServer
// * out – the socket's output to LockVerifyServer
func NewVerifyingLockFactory(lf LockFactory, in io.ByteReader, out io.ByteWriter) *VerifyingLockFactory {
	return &VerifyingLockFactory{lf: lf, in: in, out: out}
}

func (v *VerifyingLockFactory) ObtainLock(dir Directory, lockName string) (Lock, error) {
	lock, err := v.lf.ObtainLock(dir, lockName)
	if err != nil {
		return nil, err
	}
	return NewCheckedLock(v, lock), nil
}

var _ Lock = &CheckedLock{}

type CheckedLock struct {
	*VerifyingLockFactory

	lock Lock

	buff []byte
}

func NewCheckedLock(factory *VerifyingLockFactory, lock Lock) *CheckedLock {
	return &CheckedLock{
		VerifyingLockFactory: factory,
		lock:                 lock,
		buff:                 make([]byte, 1),
	}
}

func (c *CheckedLock) Close() error {
	if err := c.lock.EnsureValid(); err != nil {
		return err
	}
	return c.verify(MSG_LOCK_RELEASED)
}

func (c *CheckedLock) EnsureValid() error {
	return c.lock.EnsureValid()
}

func (c *CheckedLock) verify(message byte) error {
	err := c.out.WriteByte(message)
	if err != nil {
		return err
	}

	ret, err := c.in.ReadByte()
	if err != nil {
		return err
	}
	if ret < 0 {
		return errors.New("lock server died because of locking error")
	}
	if ret != message {
		return errors.New("protocol violation")
	}
	return nil
}
