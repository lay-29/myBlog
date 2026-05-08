package models

func NormalizeVisibility(v Visibility) Visibility {
	switch v {
	case VisibilityPublic, VisibilityTeam, VisibilityPrivate:
		return v
	default:
		return VisibilityTeam
	}
}

func NormalizeTeamRole(role TeamRole) TeamRole {
	switch role {
	case TeamRoleOwner, TeamRoleAdmin, TeamRoleMember:
		return role
	default:
		return TeamRoleMember
	}
}

func CanManageTeam(role TeamRole) bool {
	return role == TeamRoleOwner || role == TeamRoleAdmin
}

func CanWriteSpace(role TeamRole) bool {
	return role == TeamRoleOwner || role == TeamRoleAdmin || role == TeamRoleMember
}

func CanViewArticle(user *User, role TeamRole, visibility Visibility, authorID int64) bool {
	if user != nil && user.IsAdmin {
		return true
	}
	switch NormalizeVisibility(visibility) {
	case VisibilityPublic:
		return true
	case VisibilityTeam:
		return role == TeamRoleOwner || role == TeamRoleAdmin || role == TeamRoleMember
	case VisibilityPrivate:
		return user != nil && user.ID == authorID
	default:
		return false
	}
}
