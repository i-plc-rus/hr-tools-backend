package models

import (
	"slices"

	"github.com/pkg/errors"
)

type Experience string

const (
	ExperienceNoMatter   Experience = "noMatter"   // Без опыта
	ExperienceMoreThan1  Experience = "moreThan1"  // Более 1 года
	ExperienceMoreThan3  Experience = "moreThan3"  // Более 3 лет
	ExperienceMoreThan5  Experience = "moreThan5"  // Более 5 лет
	ExperienceMoreThan10 Experience = "moreThan10" // Более 10 лет
)

func ExperienceNameSlice() []string {
	return []string{"Без опыта", "Более 1 года", "Более 3 лет", "Более 5 лет", "Более 10 лет"}
}

func ExperienceSlice() []Experience {
	return []Experience{ExperienceNoMatter, ExperienceMoreThan1, ExperienceMoreThan3, ExperienceMoreThan5, ExperienceMoreThan10}
}

func (v Experience) Code() string {
	return string(v)
}

func (v Experience) Name() string {
	return v.ToString()
}

func (v Experience) Validate(optional bool) error {
	if v == "" {
		if optional {
			return nil
		}
		return errors.New("требуемый опыт не указана")
	}
	if !slices.Contains(ExperienceSlice(), v) {
		return errors.New("опыт указан некорректно")
	}
	return nil
}

func (s Experience) ToString() string {
	switch s {
	case ExperienceNoMatter:
		return "Без опыта"
	case ExperienceMoreThan1:
		return "Более 1 года"
	case ExperienceMoreThan3:
		return "Более 3 лет"
	case ExperienceMoreThan5:
		return "Более 5 лет"
	case ExperienceMoreThan10:
		return "Более 10 лет"
	}
	return ""
}
func ExperienceFromDescr(s string) Experience {
	switch s {
	case "Без опыта":
		return ExperienceNoMatter
	case "Более 1 года":
		return ExperienceMoreThan1
	case "Более 3 лет":
		return ExperienceMoreThan3
	case "Более 5 лет":
		return ExperienceMoreThan5
	case "Более 10 лет":
		return ExperienceMoreThan10
	}
	return ""
}

func (s Experience) ToPoint() int {
	switch s {
	case ExperienceNoMatter:
		return 1
	case ExperienceMoreThan1:
		return 2
	case ExperienceMoreThan3:
		return 3
	case ExperienceMoreThan5:
		return 4
	case ExperienceMoreThan10:
		return 5
	}
	return 0
}
func (s Experience) ToHHId() string {
	switch s {
	case ExperienceNoMatter:
		return "noExperience"
	case ExperienceMoreThan1:
		return "between1And3"
	case ExperienceMoreThan3:
		return "between3And6"
	case ExperienceMoreThan5:
		return "moreThan6"
	case ExperienceMoreThan10:
		return "moreThan6"
	}
	return "noExperience"
}
