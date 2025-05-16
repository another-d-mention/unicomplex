package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/binary"
	"io"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type User struct {
	id           uuid.UUID
	username     string
	passwordHash [32]byte
	passwordSalt [16]byte
	groups       []uuid.UUID

	groupNames []string
}

func (u *User) ID() uuid.UUID {
	id, _ := uuid.FromBytes(u.id[:])
	return id
}

func (u *User) Username() string {
	return u.username
}

func (u *User) GroupIDs() []uuid.UUID {
	list := make([]uuid.UUID, len(u.groups))
	copy(list, u.groups)
	return list
}

func (u *User) GroupNames() []string {
	list := make([]string, len(u.groupNames))
	copy(list, u.groupNames)
	return list
}

func (u *User) marshalBinary(w io.Writer) {
	_ = binary.Write(w, binary.BigEndian, u.id[:])
	_ = binary.Write(w, binary.BigEndian, byte(len(u.username)))
	_ = binary.Write(w, binary.BigEndian, []byte(u.username))
	_ = binary.Write(w, binary.BigEndian, u.passwordHash[:])
	_ = binary.Write(w, binary.BigEndian, u.passwordSalt[:])
	_ = binary.Write(w, binary.BigEndian, byte(len(u.groups)))
	for _, g := range u.groups {
		_ = binary.Write(w, binary.BigEndian, g[:])
	}
}

func (u *User) unmarshalBinary(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &u.id); err != nil {
		return err
	}
	var usernameLen byte
	if err := binary.Read(r, binary.BigEndian, &usernameLen); err != nil {
		return err
	}
	var usernameBytes = make([]byte, int(usernameLen))
	if err := binary.Read(r, binary.BigEndian, usernameBytes); err != nil {
		return err
	}
	u.username = string(usernameBytes)
	if err := binary.Read(r, binary.BigEndian, &u.passwordHash); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &u.passwordSalt); err != nil {
		return err
	}
	var groupCount byte
	if err := binary.Read(r, binary.BigEndian, &groupCount); err != nil {
		return err
	}
	for i := 0; i < int(groupCount); i++ {
		var g uuid.UUID
		if err := binary.Read(r, binary.BigEndian, &g); err != nil {
			return err
		}
		u.groups = append(u.groups, g)
	}
	return nil
}

func newUser(username, password string) (*User, error) {
	id, _ := uuid.NewRandom()
	var salt [16]byte
	_, err := rand.Read(salt[:])
	if err != nil {
		return nil, err
	}

	u := &User{
		id:           id,
		username:     username,
		passwordSalt: salt,
		groups:       []uuid.UUID{},
	}

	u.passwordHash = u.hashPassword(password)
	return u, nil
}

func (u *User) hashPassword(password string) [32]byte {
	hashed := argon2.IDKey([]byte(password), u.passwordSalt[:], 1, 64*1024, 4, 32)
	var pwd [32]byte
	copy(pwd[:], hashed)
	return pwd
}

func (u *User) HasGroup(groupNameOrId string) bool {
	for _, g := range u.groupNames {
		if g == groupNameOrId {
			return true
		}
	}
	_, err := uuid.Parse(groupNameOrId)
	if err != nil {
		return false
	}
	for _, g := range u.groups {
		if g.String() == groupNameOrId {
			return true
		}
	}
	return false
}

func (u *User) VerifyPassword(password string) bool {
	p := u.hashPassword(password)
	return hmac.Equal(u.passwordHash[:], p[:])
}

type Group struct {
	id   uuid.UUID
	name string

	userIds   []uuid.UUID
	userNames []string
}

func (g *Group) ID() uuid.UUID {
	id, _ := uuid.FromBytes(g.id[:])
	return id
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) UserIDs() []uuid.UUID {
	list := make([]uuid.UUID, len(g.userIds))
	copy(list, g.userIds)
	return list
}

func (g *Group) UserNames() []string {
	list := make([]string, len(g.userNames))
	copy(list, g.userNames)
	return list
}

func newGroup(name string) *Group {
	id, _ := uuid.NewRandom()
	return &Group{
		id:   id,
		name: name,
	}
}

func (g *Group) marshalBinary(w io.Writer) {
	_ = binary.Write(w, binary.BigEndian, g.id[:])
	_ = binary.Write(w, binary.BigEndian, byte(len(g.name)))
	_ = binary.Write(w, binary.BigEndian, []byte(g.name))
}

func (g *Group) unmarshalBinary(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &g.id); err != nil {
		return err
	}
	var nameLen byte
	if err := binary.Read(r, binary.BigEndian, &nameLen); err != nil {
		return err
	}
	var nameBytes = make([]byte, int(nameLen))
	if err := binary.Read(r, binary.BigEndian, nameBytes); err != nil {
		return err
	}
	g.name = string(nameBytes)
	return nil
}
