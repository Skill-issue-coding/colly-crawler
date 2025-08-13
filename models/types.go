package models

type Literature struct {
	// More to come
}

type Examinations struct {
	// More to come
}

type Plan struct {
	// More to come
}

type Overview struct {
	Subject        string `json:"main_subject"`
	Level          string `json:"education_level"`
	Type           string `json:"course_type"`
	Examiner       string `json:"examiner"`
	Director       string `json:"study_director"`
	ScheduledHours string `json:"scheduled_hours"`
	SelfStudyHours string `json:"self_study_hours"`
	Period         string `json:"period"`
	Block          string `json:"block"`
	Language       string `json:"language"`
	Campus         string `json:"campus"`
	VOF            string `json:"vof"`
}

type Course struct {
	Name         string         `json:"name"`
	Code         string         `json:"course_code"`
	Credits      string         `json:"credits"`
	Url          string         `json:"url"`
	Overview     Overview       `json:"overview"`
	Plan         Plan           `json:"course_plan"`
	Examinations []Examinations `json:"examinations"`
	Literature   []Literature   `json:"literature"`
}

type Semester struct {
	Name    string   `json:"name"`
	Courses []Course `json:"courses"`
}

type Program struct {
	Name      string     `json:"name"`
	Credits   string     `json:"credits"`
	Url       string     `json:"url"`
	Semesters []Semester `json:"semesters"`
}
