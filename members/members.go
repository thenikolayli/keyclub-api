package members

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// struct to represent a member
type Member struct {
	ID            string  `db:"id"`
	Firstname     string  `db:"first_name"`
	Nickname      string  `db:"nickname"`
	Middlename    string  `db:"middle_name"`
	Lastname      string  `db:"last_name"`
	AllHours      float64 `db:"all_hours"`
	TermHours     float64 `db:"term_hours"`
	GradYear      int     `db:"grad_year"`
	Class         string  `db:"class"`
	Strikes       int     `db:"strikes"`
	PersonalEmail string  `db:"personal_email"`
	SchoolEmail   string  `db:"school_email"`
	PhoneNumber   string  `db:"phone_number"`
	ShirtSize     string  `db:"shirt_size"`
	PaidDues      bool    `db:"paid_dues"`
}

// For functions like member lookup
type FormattedMember struct {
	Name          string
	AllHours      float64
	TermHours     float64
	GradYear      int
	Class         string
	Strikes       int
	PersonalEmail string
	SchoolEmail   string
	PhoneNumber   string
	ShirtSize     string
	PaidDues      bool
}

// struct to represnt a formatted name
type Name struct {
	First  string
	Nick   string
	Middle string
	Last   string
}

// formats a member for a more readable output, such as for the member lookup command
func (member Member) Format() FormattedMember {
	name := cases.Title(language.English).String(member.Firstname)
	if member.Nickname != "" {
		name += cases.Title(language.English).String(fmt.Sprintf(` "%v" `, member.Nickname))
	}
	if member.Middlename != "" {
		name += " " + cases.Title(language.English).String(member.Middlename)
	}
	name += " " + cases.Title(language.English).String(member.Lastname)

	if member.PersonalEmail == "" {
		member.PersonalEmail = "N/A"
	}
	if member.SchoolEmail == "" {
		member.SchoolEmail = "N/A"
	}
	if member.PhoneNumber == "" {
		member.PhoneNumber = "N/A"
	} else {
		member.PhoneNumber = formatPhoneNumber(member.PhoneNumber)
	}
	if member.ShirtSize == "" {
		member.ShirtSize = "N/A"
	}
	return FormattedMember{
		Name:          name,
		AllHours:      member.AllHours,
		TermHours:     member.TermHours,
		GradYear:      member.GradYear,
		Class:         member.Class,
		Strikes:       member.Strikes,
		PersonalEmail: member.PersonalEmail,
		SchoolEmail:   member.SchoolEmail,
		PhoneNumber:   member.PhoneNumber,
		ShirtSize:     member.ShirtSize,
		PaidDues:      member.PaidDues,
	}
}

// formats phone numbers into this standard format: (XXX) XXX-XXXX
func formatPhoneNumber(phoneNumber string) string {
	cleanNumber := strings.ReplaceAll(phoneNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")

	if len(cleanNumber) == 10 {
		return fmt.Sprintf("(%s) %s-%s", cleanNumber[0:3], cleanNumber[3:6], cleanNumber[6:10])
	} else {
		return phoneNumber // if the number isn't 10 digits, return it as is
	}
}

// creates a new instance of type Name based on a string input
// this is a standalone function because it's often called on a string input, not a member struct
// names are in a First "Nick" Middle Last format
func NewName(name string) Name {
	nameParts := strings.Fields(name)

	if len(nameParts) == 2 {
		return Name{
			First: strings.ToLower(strings.Trim(nameParts[0], `"`)),
			Last:  strings.ToLower(strings.Trim(nameParts[1], `"`)),
		}
	} else if len(nameParts) == 3 {
		// First "Nick" Last vs First Middle Last
		if strings.Contains(nameParts[1], `"`) {
			return Name{
				First: strings.ToLower(strings.Trim(nameParts[0], `"`)),
				Nick:  strings.ToLower(strings.Trim(nameParts[1], `"`)),
				Last:  strings.ToLower(strings.Trim(nameParts[2], `"`)),
			}
		} else {
			return Name{
				First:  strings.ToLower(strings.Trim(nameParts[0], `"`)),
				Middle: strings.ToLower(strings.Trim(nameParts[1], `"`)),
				Last:   strings.ToLower(strings.Trim(nameParts[2], `"`)),
			}
		}
	} else if len(nameParts) == 4 {
		return Name{
			First:  strings.ToLower(strings.Trim(nameParts[0], `"`)),
			Nick:   strings.ToLower(strings.Trim(nameParts[1], `"`)),
			Middle: strings.ToLower(strings.Trim(nameParts[2], `"`)),
			Last:   strings.ToLower(strings.Trim(nameParts[3], `"`)),
		}
	}
	return Name{
		First: strings.ToLower(strings.Trim(nameParts[0], `"`)),
		Nick:  strings.ToLower(strings.Trim(nameParts[0], `"`)),
	}
}

// name speaks for itself
func SameName(name1 Name, name2 Name) bool {
	if name1.First == name2.First && name1.Last == name2.Last {
		return true
	}
	if name1.First == name2.Last && name1.Last == name2.First {
		return true
	}
	if name1.Nick != "" && (name1.Nick == name2.First || name1.Nick == name2.Nick) {
		return true
	}
	return false
}
