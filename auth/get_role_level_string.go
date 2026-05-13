package auth

func GetRoleLevelString(roleLevel int) string {
	switch roleLevel {
	case 0:
		return "member"
	case 1:
		return "leader"
	case 2:
		return "officer"
	default:
		return "member"
	}
}
