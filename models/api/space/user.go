package spaceapimodels

import (
	"errors"
	"hr-tools-backend/models"
	apimodels "hr-tools-backend/models/api"
	"slices"
	"time"
)

type CreateUser struct {
	Password string `json:"password"`
	SpaceUserCommonData
}

type UpdateUser struct {
	Password *string `json:"password"`
	SpaceUserUpdateData
}

type SpaceUser struct {
	ID string `json:"id"`
	SpaceUserCommonData
	IsEmailVerified bool   `json:"is_email_verified"` // Email подтвержден
	NewEmail        string `json:"new_email"`         // Новый email, который станет основным после подтверждения
	JobTitleName    string `json:"job_title_name"`    // Навание должности
	RoleName        string `json:"role_name"`
}

type SpaceUserExt struct {
	SpaceUser
	LicenseStatus   models.LicenseStatus                  `json:"license_status"`    // Статус лицензии
	LicenseReadOnly bool                                  `json:"license_read_only"` // Блокировка мутаций
	Permissions     map[models.Module][]models.Permission `json:"permissions"`
}

type SpaceUserCommonData struct {
	SpaceID         string          `json:"space_id"`
	Email           string          `json:"email"` // Email пользователя
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	PhoneNumber     string          `json:"phone_number"`
	Role            models.UserRole `json:"role"`
	TextSign        string          `json:"text_sign"`         // Текст подписи
	JobTitleID      string          `json:"job_title_id"`      // Идентификатор должности
	Status          string          `json:"status"`            // Статус пользователя
	StatusChangedAt time.Time       `json:"status_changed_at"` // Дата изменения статуса
	StatusComment   *string         `json:"status_comment"`    // Комментарий к статусу
}

type SpaceUserUpdateData struct {
	SpaceID     string  `json:"space_id"`
	Email       string  `json:"email"` // Email пользователя
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	PhoneNumber string  `json:"phone_number"`
	Role        *string `json:"role"`
	TextSign    *string `json:"text_sign"`    // Текст подписи
	JobTitleID  *string `json:"job_title_id"` // Идентификатор должности
}

func (r SpaceUserUpdateData) Validate() error {
	if r.Email == "" {
		return errors.New("не указан емайл")
	}
	if r.PhoneNumber == "" {
		return errors.New("не указан телефон")
	}
	if r.FirstName == "" {
		return errors.New("не указано имя")
	}
	if r.LastName == "" {
		return errors.New("не указана фамилия")
	}
	if r.Role != nil && !slices.Contains(models.AllAvailableRoles, models.UserRole(*r.Role)) {
		return errors.New("указаная роль отсутсвует")
	}
	return nil
}

func (r CreateUser) Validate() error {
	if r.Password == "" {
		return errors.New("не указан пароль")
	}
	return r.SpaceUserCommonData.Validate()
}

func (r SpaceUserCommonData) Validate() error {
	if r.Email == "" {
		return errors.New("не указан емайл")
	}
	if r.PhoneNumber == "" {
		return errors.New("не указан телефон")
	}
	if r.FirstName == "" {
		return errors.New("не указано имя")
	}
	if r.LastName == "" {
		return errors.New("не указана фамилия")
	}
	if !slices.Contains(models.AllAvailableRoles, models.UserRole(r.Role)) {
		return errors.New("указаная роль отсутсвует")
	}
	return nil
}

type SpaceUserProfileData struct {
	Email               string  `json:"email"`                 // Email пользователя
	FirstName           string  `json:"first_name"`            // Имя
	LastName            string  `json:"last_name"`             // Фамилия
	PhoneNumber         string  `json:"phone_number"`          // Телефон
	InternalPhoneNumber string  `json:"internal_phone_number"` // Внутренний номер
	UsePersonalSign     bool    `json:"use_personal_sign"`     // Персональная подпись
	TextSign            string  `json:"text_sign"`             // Текст подписи
	JobTitleID          *string `json:"job_title_id"`          // Идентификатор должности
}

func (r SpaceUserProfileData) Validate() error {
	if r.Email == "" {
		return errors.New("не указан емайл")
	}
	if r.FirstName == "" {
		return errors.New("не указано имя")
	}
	return nil
}

type SpaceUserProfileView struct {
	SpaceUserProfileData
	ID              string          `json:"id"`   // Идентфикатор пользователя
	Role            models.UserRole `json:"role"` // Роль
	RoleName        string          `json:"role_name"`
	IsEmailVerified bool            `json:"is_email_verified"` // Email подтвержден
	NewEmail        string          `json:"new_email"`         // Новый email, который станет основным после подтверждения
	JobTitleName    string          `json:"job_title_name"`    // Должность
}

type PasswordChange struct {
	CurrentPassword string `json:"current_password"` // Текущий пароль
	NewPassword     string `json:"new_password"`     // Новый пароль
}

func (r PasswordChange) Validate() error {
	if r.CurrentPassword == "" {
		return errors.New("Не указан текущий пароль")
	}
	if r.NewPassword == "" {
		return errors.New("Не указан новый пароль")
	}
	return nil
}

type SpaceUserFilter struct {
	apimodels.Pagination
	Search string        `json:"search"` // Поиск
	Sort   SpaceUserSort `json:"sort"`   // Сортировка
	Status *string       `json:"status"` // Статус пользователя
}

type SpaceUserSort struct {
	NameDesc  *bool `json:"fio_desc"`   // Имя, порядок сортировки false = ASC/ true = DESC / nil = нет
	EmailDesc *bool `json:"email_desc"` // Email, порядок сортировки false = ASC/ true = DESC / nil = нет
	RoleDesc  *bool `json:"role_desc"`  // Роль добавления, порядок сортировки false = ASC/ true = DESC / nil = нет
}

type UpdateUserStatus struct {
	Status  string  `json:"status"`  // Статус пользователя: WORKING, VACATION, DISMISSED
	Comment *string `json:"comment"` // Комментарий к статусу (опционально)
}

func (r UpdateUserStatus) Validate() error {
	if r.Status == "" {
		return errors.New("не указан статус")
	}
	validStatuses := map[string]bool{
		"WORKING":   true,
		"VACATION":  true,
		"DISMISSED": true,
	}
	if !validStatuses[r.Status] {
		return errors.New("неверный статус. Допустимые значения: WORKING, VACATION, DISMISSED")
	}
	return nil
}
