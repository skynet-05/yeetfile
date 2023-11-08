package mail

import "text/template"

type ResetEmail struct {
	Code   string
	Email  string
	Domain string
}

var resetSubject = "YeetFile Password Reset"
var resetBodyTemplate = template.Must(template.New("").Parse(
	"Your YeetFile password reset code is {{.Code}}.\n\n" +
		"Enter this code on the reset page, or use the link " +
		"below to reset your password.\n\n" +
		"{{.Domain}}/reset?email={{.Email}}&code={{.Code}}"))
