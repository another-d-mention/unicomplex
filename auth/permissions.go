package auth

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/google/uuid"
)

type Permission uint64

func (p Permission) SetAll() Permission {
	return math.MaxUint64
}

func (p Permission) Clear() Permission {
	return Permission(0)
}

func (p Permission) Enable(permissions uint64) Permission {
	val := uint64(p)
	val |= permissions
	return Permission(val)
}

func (p Permission) Disable(permissions uint64) Permission {
	val := uint64(p)
	val &= ^permissions
	return Permission(val)
}

func (p Permission) HasPermission(permission uint64) bool {
	val := uint64(p)
	val &= permission
	return val != 0
}

func (p Permission) HasAnyPermission(list ...uint64) bool {
	for _, o := range list {
		if p.HasPermission(o) {
			return true
		}
	}
	return false
}

func (p Permission) HasAllPermissions(list ...uint64) bool {
	for _, o := range list {
		if !p.HasPermission(o) {
			return false
		}
	}
	return true
}

func (p Permission) String() string {
	return fmt.Sprintf("%064b", p)
}

type ACL struct {
	lock   sync.RWMutex
	id     uuid.UUID
	users  map[uuid.UUID]Permission
	groups map[uuid.UUID]Permission
}

func NewACL() *ACL {
	return &ACL{
		id:     uuid.New(),
		users:  make(map[uuid.UUID]Permission),
		groups: make(map[uuid.UUID]Permission),
	}
}

func (a *ACL) ID() uuid.UUID {
	return a.id
}

func (a *ACL) MarshalBinary(w io.Writer) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	_ = binary.Write(w, binary.BigEndian, uint16(len(a.users)))
	for id, p := range a.users {
		_ = binary.Write(w, binary.BigEndian, id[:])
		_ = binary.Write(w, binary.BigEndian, uint64(p))
	}

	_ = binary.Write(w, binary.BigEndian, uint16(len(a.groups)))
	for id, p := range a.groups {
		_ = binary.Write(w, binary.BigEndian, id[:])
		_ = binary.Write(w, binary.BigEndian, uint64(p))
	}

	return nil
}

func (a *ACL) UnmarshalBinary(r io.Reader) error {
	var numUsers, numGroups uint16
	err := binary.Read(r, binary.BigEndian, &numUsers)
	if err != nil {
		return err
	}
	users := make(map[uuid.UUID]Permission, numUsers)
	for i := 0; i < int(numUsers); i++ {
		var id uuid.UUID
		err := binary.Read(r, binary.BigEndian, &id)
		if err != nil {
			return err
		}
		var p Permission
		err = binary.Read(r, binary.BigEndian, &p)
		if err != nil {
			return err
		}
		users[id] = p
	}

	err = binary.Read(r, binary.BigEndian, &numGroups)
	if err != nil {
		return err
	}
	groups := make(map[uuid.UUID]Permission, numGroups)
	for i := 0; i < int(numGroups); i++ {
		var id uuid.UUID
		err := binary.Read(r, binary.BigEndian, &id)
		if err != nil {
			return err
		}
		var p Permission
		err = binary.Read(r, binary.BigEndian, &p)
		if err != nil {
			return err
		}
		groups[id] = p
	}

	a.lock.Lock()
	a.users = users
	a.groups = groups
	a.lock.Unlock()
	return nil
}
