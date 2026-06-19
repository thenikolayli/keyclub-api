package google

type MembersSheetRangesType struct {
	SheetName     string
	Names         string
	AllHours      string
	TermHours     string
	GradYear      string
	Class         string
	Strikes       string
	PersonalEmail string
	SchoolEmail   string
	PhoneNumber   string
	ShirtSizes    string
	PaidDues      string
}

type EventsSheetRangesType struct {
	SheetName     string
	Events        string
	Dates         string
	StartTimes    string
	EndTimes      string
	Addresses     string
	NofSlots      string
	NofVolunteers string
	TotalHours    string
	Leaders       string
	MadeBy        string
}

type EventsMembersSheetRangesType struct {
	SheetName string
	Events    string
	Members   string
}
