package auth

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/google/uuid"
)

type Auth struct {
	lock   sync.RWMutex
	users  map[string]*User
	groups map[string]*Group
}

func New() *Auth {
	return &Auth{
		users:  make(map[string]*User),
		groups: make(map[string]*Group),
	}
}

// ---------------- USERS ---------------

func (a *Auth) Users() []*User {
	a.lock.RLock()
	defer a.lock.RUnlock()

	list := make([]*User, 0, len(a.users))
	for _, u := range a.users {
		list = append(list, u)
	}
	return list
}

func (a *Auth) AddUser(username, password string) error {
	if _, ok := a.GetUser(username); ok {
		return ErrUserAlreadyExists
	}

	u, err := newUser(username, password)
	if err != nil {
		return err
	}

	a.lock.Lock()
	a.users[u.username] = u
	a.lock.Unlock()

	return nil
}

func (a *Auth) GetUser(usernameOrId string) (*User, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	u, ok := a.users[usernameOrId]
	if ok {
		return u, true
	}

	_, err := uuid.Parse(usernameOrId)
	if err == nil {
		for _, u = range a.users {
			if u.id.String() == usernameOrId {
				return u, true
			}
		}
	}

	return nil, false
}

func (a *Auth) DeleteUser(usernameOrId string) error {
	u, ok := a.GetUser(usernameOrId)
	if !ok {
		return ErrUserNotFound
	}

	for _, g := range u.groupNames {
		if err := a.RemoveUserFromGroup(u.username, g); err != nil {
			return err
		}
	}

	a.lock.Lock()
	delete(a.users, u.username)
	a.lock.Unlock()

	return nil
}

// ---------------- GROUPS ---------------

func (a *Auth) Groups() []*Group {
	a.lock.RLock()
	defer a.lock.RUnlock()

	list := make([]*Group, 0, len(a.groups))
	for _, g := range a.groups {
		list = append(list, g)
	}
	return list
}

func (a *Auth) AddGroup(name string) error {
	if _, ok := a.GetGroup(name); ok {
		return ErrGroupAlreadyExists
	}

	g := newGroup(name)

	a.lock.Lock()
	a.groups[g.name] = g
	a.lock.Unlock()
	return nil
}

func (a *Auth) GetGroup(nameOrId string) (*Group, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	g, ok := a.groups[nameOrId]
	if ok {
		return g, true
	}
	_, err := uuid.Parse(nameOrId)
	if err == nil {
		for _, g = range a.groups {
			if g.id.String() == nameOrId {
				return g, true
			}
		}
	}
	return nil, false
}

func (a *Auth) DeleteGroup(name string) error {
	g, ok := a.GetGroup(name)
	if !ok {
		return ErrGroupNotFound
	}

	for _, username := range g.UserNames() {
		if err := a.RemoveUserFromGroup(username, name); err != nil {
			return err
		}
	}

	a.lock.Lock()
	delete(a.groups, name)
	a.lock.Unlock()
	return nil

}

func (a *Auth) AddUserToGroup(username, groupName string) error {
	u, ok := a.GetUser(username) // validate user
	if !ok {
		return ErrUserNotFound
	}

	g, ok := a.GetGroup(groupName) // validate group
	if !ok {
		return ErrGroupNotFound
	}

	if u.HasGroup(groupName) { // already in the group
		return nil
	}

	g.userIds = append(g.userIds, u.id)
	g.userNames = append(g.userNames, u.username)

	u.groups = append(u.groups, g.id)
	u.groupNames = append(u.groupNames, g.name)
	return nil
}

func (a *Auth) RemoveUserFromGroup(username, groupName string) error {
	u, ok := a.GetUser(username) // validate user
	if !ok {
		return ErrUserNotFound
	}

	g, ok := a.GetGroup(groupName) // validate group
	if !ok {
		return ErrGroupNotFound
	}

	if !u.HasGroup(groupName) { // not in the group
		return nil
	}

	for i, id := range u.groups {
		if id == g.id {
			u.groups = append(u.groups[:i], u.groups[i+1:]...)
			u.groupNames = append(u.groupNames[:i], u.groupNames[i+1:]...)
			break
		}
	}

	for i, id := range g.userIds {
		if id == u.id {
			g.userIds = append(g.userIds[:i], g.userIds[i+1:]...)
			g.userNames = append(g.userNames[:i], g.userNames[i+1:]...)
			break
		}
	}

	return nil
}

// ---------------- ACL ---------------

func (a *Auth) ACLUsers(acl *ACL) map[string]Permission {
	acl.lock.RLock()
	defer acl.lock.RUnlock()

	list := make(map[string]Permission, len(acl.users))
	for u, p := range acl.users {
		usr, ok := a.GetUser(u.String())
		if ok {
			list[usr.username] = p
		}
	}
	return list
}

func (a *Auth) AddUserToACL(acl *ACL, usernameOrId string, permission Permission) error {
	u, ok := a.GetUser(usernameOrId) // validate user
	if !ok {
		return ErrUserNotFound
	}
	acl.lock.Lock()
	acl.users[u.id] = permission
	acl.lock.Unlock()
	return nil
}

func (a *Auth) RemoveUserFromACL(acl *ACL, usernameOrId string) error {
	u, ok := a.GetUser(usernameOrId) // validate user
	if !ok {
		return ErrUserNotFound
	}

	acl.lock.Lock()
	delete(acl.users, u.id)
	acl.lock.Unlock()

	return nil
}

func (a *Auth) ACLGroups(acl *ACL) map[string]Permission {
	acl.lock.RLock()
	defer acl.lock.RUnlock()

	list := make(map[string]Permission, len(acl.groups))
	for g, p := range acl.groups {
		gr, ok := a.GetGroup(g.String())
		if ok {
			list[gr.name] = p
		}
	}
	return list
}

func (a *Auth) GetPermission(acl *ACL, nameOrId string) (Permission, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if user, ok := a.GetUser(nameOrId); ok { // it's a user id or name
		userPermission, k := acl.users[user.id] // we have ACL permissions for this user
		if k {
			return userPermission, true
		}

		// we look in the groups the user belongs to and check their permissions
		groups := user.GroupIDs()
		maximum := Permission(0)
		for _, id := range groups {
			if groupPermission, k := acl.groups[id]; k {
				if groupPermission > maximum {
					maximum = groupPermission
				}
			}
		}

		if maximum == 0 {
			return Permission(0), false
		}

		return maximum, true
	}

	if group, ok := a.GetGroup(nameOrId); ok { // it's a group id or name
		groupPermission, k := acl.groups[group.id] // we have ACL permissions for this group
		if k {
			return groupPermission, true
		}
	}

	return Permission(0), false
}

func (a *Auth) AddGroupToACL(acl *ACL, groupOrId string, permission Permission) error {
	g, ok := a.GetGroup(groupOrId) // validate group
	if !ok {
		return ErrGroupNotFound
	}

	acl.lock.Lock()
	acl.groups[g.id] = permission
	acl.lock.Unlock()

	return nil
}

func (a *Auth) RemoveGroupFromACL(acl *ACL, groupOrId string) error {
	g, ok := a.GetGroup(groupOrId) // validate group
	if !ok {
		return ErrGroupNotFound
	}

	acl.lock.Lock()
	delete(acl.groups, g.id)
	acl.lock.Unlock()

	return nil
}

// ---------------- MISC ---------------

func (a *Auth) MarshalBinary(w io.Writer) error {
	a.lock.RLock()
	defer a.lock.RUnlock()

	_ = binary.Write(w, binary.BigEndian, uint64(len(a.users)))
	for _, u := range a.users {
		u.marshalBinary(w)
	}
	_ = binary.Write(w, binary.BigEndian, uint64(len(a.groups)))
	for _, g := range a.groups {
		g.marshalBinary(w)
	}
	return nil
}

func (a *Auth) UnmarshalBinary(r io.Reader) error {
	var numUsers, numGroups byte
	err := binary.Read(r, binary.BigEndian, &numUsers)
	if err != nil {
		return err
	}
	users := make(map[string]*User, numUsers)
	for i := 0; i < int(numUsers); i++ {
		u := new(User)
		if err = u.unmarshalBinary(r); err != nil {
			return err
		}
		users[u.username] = u
	}

	err = binary.Read(r, binary.BigEndian, &numGroups)
	if err != nil {
		return err
	}

	groups := make(map[string]*Group, numGroups)
	for i := 0; i < int(numGroups); i++ {
		g := new(Group)
		if err = g.unmarshalBinary(r); err != nil {
			return err
		}
		groups[g.name] = g
	}

	a.lock.Lock()
	a.users = users
	a.groups = groups
	a.lock.Unlock()

	return a.load()
}

func (a *Auth) load() error {
	a.lock.Lock()
	defer a.lock.Unlock()

	for _, u := range a.users {
		for _, id := range u.groups {
			g, ok := a.GetGroup(id.String())
			if !ok {
				continue
			}

			g.userIds = append(g.userIds, u.id)
			g.userNames = append(g.userNames, u.username)

			u.groupNames = append(u.groupNames, g.name)
		}
	}

	return nil
}
