package auth

import (
	"testing"
)

type Perms int

const (
	CanRead    Perms = 1 << iota // 00000001
	CanWrite                     // 00000010
	CanDelete                    // 00000100
	CanExecute                   // 00001000
)

func TestACL(t *testing.T) {
	auth := New()
	if err := auth.AddUser("user1", "testpassword"); err != nil {
		t.Fatal(err)
	}
	if err := auth.AddUser("user2", "testpassword"); err != nil {
		t.Fatal(err)
	}
	if err := auth.AddUser("user3", "testpassword"); err != nil {
		t.Fatal(err)
	}

	if err := auth.AddGroup("group"); err != nil {
		t.Fatal(err)
	}
	if err := auth.AddUserToGroup("user2", "group"); err != nil {
		t.Fatal(err)
	}
	if err := auth.AddUserToGroup("user3", "group"); err != nil {
		t.Fatal(err)
	}

	acl := NewACL()

	if err := auth.AddUserToACL(acl, "user1", Permission(CanExecute|CanWrite)); err != nil {
		t.Fatal(err)
	}
	if err := auth.AddGroupToACL(acl, "group", Permission(CanRead)); err != nil {
		t.Fatal(err)
	}
	if err := auth.AddUserToACL(acl, "user3", Permission(CanRead|CanWrite)); err != nil {
		t.Fatal(err)
	}

	if len(auth.ACLUsers(acl)) != 2 {
		t.Error("Expected 2 user in ACL, got", len(auth.ACLUsers(acl)))
	}
	if len(auth.ACLGroups(acl)) != 1 {
		t.Error("Expected 1 group in ACL, got", len(auth.ACLGroups(acl)))
	}

	p, ok := auth.GetPermission(acl, "user1")
	if !ok {
		t.Error("User not found in ACL")
	}
	if p != Permission(CanExecute|CanWrite) {
		t.Error("Expected user permission, got", p)
	}

	p, ok = auth.GetPermission(acl, "user2")
	if !ok {
		t.Error("User not found in ACL")
	}
	if p != Permission(CanRead) {
		t.Error("Expected user permission, got", p)
	}

	p, ok = auth.GetPermission(acl, "user3")
	if !ok {
		t.Error("User not found in ACL")
	}
	if !p.HasPermission(uint64(CanWrite)) {
		t.Error("Expected true", p)
	}
	p.Disable(uint64(CanWrite))
	p.Enable(uint64(CanWrite))
	if !p.HasPermission(uint64(CanWrite)) {
		t.Error("Expected true", p)
	}
}

func TestUser(t *testing.T) {
	u, err := newUser("testuser", "testpassword")
	if err != nil {
		t.Fatal(err)
	}
	if u.VerifyPassword("asd") {
		t.Error("Password should failed")
	}
	if !u.VerifyPassword("testpassword") {
		t.Error("Password verification failed")
	}
}

func TestGroup(t *testing.T) {
	auth := New()
	err := auth.AddUser("testuser", "testpassword")
	if err != nil {
		t.Fatal(err)
	}
	err = auth.AddUser("testuser", "testpassword")
	if err == nil {
		t.Error("User already exists")
	}
	err = auth.AddGroup("testgroup")
	if err != nil {
		t.Fatal(err)
	}
	err = auth.AddGroup("testgroup")
	if err == nil {
		t.Error("Group already exists")
	}
	err = auth.AddUserToGroup("testuser", "testgroup")
	if err != nil {
		t.Fatal(err)
	}
	err = auth.AddUserToGroup("testuser", "missing")
	if err == nil {
		t.Error("invalid group")
	}
	err = auth.AddUserToGroup("tmp", "testgroup")
	if err == nil {
		t.Error("invalid user")
	}
	err = auth.AddUserToGroup("testuser", "testgroup")
	if err != nil {
		t.Error("should have silently failed")
	}
	err = auth.AddGroup("newgroup")
	if err != nil {
		t.Fatal(err)
	}
	err = auth.AddUserToGroup("testuser", "newgroup")
	if err != nil {
		t.Fatal(err)
	}

	u, ok := auth.GetUser("testuser")
	if !ok {
		t.Error("User not found")
	}
	if len(u.GroupNames()) != 2 {
		t.Error("User should have 2 groups")
	}
	if len(u.GroupIDs()) != 2 {
		t.Error("User should have 2 groups id")
	}
	g, ok := auth.GetGroup("testgroup")
	if !ok {
		t.Error("Group not found")
	}
	if len(g.UserNames()) != 1 {
		t.Error("Group should have 1 user")
	}

	if len(g.UserIDs()) != 1 {
		t.Error("Group should have 1 user id")
	}
	if len(auth.Groups()) != 2 {
		t.Error("Auth should have 2 groups")
	}
	if len(auth.Users()) != 1 {
		t.Error("Auth should have 1 user")
	}
}
