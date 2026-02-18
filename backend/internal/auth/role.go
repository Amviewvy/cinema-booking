package auth

var AdminEmails = map[string]bool{
	"admin1":             true,
	"youviv14@gmail.com": true,
}

func GetRoleByEmail(email string) string {
	if AdminEmails[email] {
		return "admin"
	}
	return "user"
}
