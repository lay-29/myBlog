package models

import "testing"

func TestCanViewArticle(t *testing.T) {
	author := &User{ID: 1}
	member := &User{ID: 2}
	admin := &User{ID: 3, IsAdmin: true}

	if !CanViewArticle(nil, "", VisibilityPublic, author.ID) {
		t.Fatal("public article should be visible anonymously")
	}
	if CanViewArticle(nil, "", VisibilityTeam, author.ID) {
		t.Fatal("team article should not be anonymous")
	}
	if !CanViewArticle(member, TeamRoleMember, VisibilityTeam, author.ID) {
		t.Fatal("team member should see team article")
	}
	if CanViewArticle(member, TeamRoleMember, VisibilityPrivate, author.ID) {
		t.Fatal("member should not see another author's private article")
	}
	if !CanViewArticle(author, "", VisibilityPrivate, author.ID) {
		t.Fatal("author should see private article")
	}
	if !CanViewArticle(admin, "", VisibilityPrivate, author.ID) {
		t.Fatal("admin should see private article")
	}
}
