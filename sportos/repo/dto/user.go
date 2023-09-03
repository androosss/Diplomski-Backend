package dto

import (
	"backend/sportos/repo/util"
	"fmt"
	"strings"
	"time"
)

type UserType string

const (
	UT_PLAYER UserType = "player"
	UT_COACH  UserType = "coach"
	UT_PLACE  UserType = "place"
	UT_ADMIN  UserType = "admin"
)

func (ut UserType) IsValid() bool {
	return ut == UT_ADMIN || ut == UT_COACH || ut == UT_PLACE || ut == UT_PLAYER
}

type User struct {
	Username          string     `json:"username" column:"user_id"`
	Email             string     `json:"email" column:"email"`
	EmailVerified     int        `json:"emailVerified" column:"email_verified"`
	UserType          UserType   `json:"userType" column:"userType"`
	PasswordHash      string     `json:"passwordHash" column:"passwordHash"`
	Token             *string    `json:"token" column:"token"`
	TokenValidUntil   *time.Time `json:"tokenValidUntil" column:"token_valid_until"`
	TokenRefreshUntil *time.Time `json:"tokenRefreshUntil" column:"token_refresh_until"`
	EditInfoCUD
}

func (s *User) GetTableName() SportosEntity {
	return "user"
}

func (s *User) GetId() string {
	return s.Username
}

type UserSearchParams struct {
	Username *string   `json:"username,omitempty"`
	Email    *string   `json:"email,omitempty"`
	UserType *UserType `json:"userType,omitempty"`
	Token    *string   `json:"token,omitempty"`
	EditInfoCUDSearchParams
	UserSortParams
	PagingSearchParams
	prefix string
}

func (sp *UserSearchParams) GetTablePrefix() string {
	if sp.prefix != "" {
		return sp.prefix
	}
	return "usr"
}

func (sp *UserSearchParams) SetTablePrefix(prefix string) {
	sp.prefix = prefix
	sp.UserSortParams.SetTablePrefix(prefix)
}

func (sp *UserSearchParams) validate() error {
	err := sp.EditInfoCUDSearchParams.validate()
	if err != nil {
		return err
	}
	err = sp.PagingSearchParams.validate()
	if err != nil {
		return err
	}
	return nil
}

func (sp *UserSearchParams) joinTables(query *string) {

}

func (sp *UserSearchParams) appendSearchQuery(query *string, params *[]interface{}) {
	// schedule params
	tablePrefix := sp.GetTablePrefix()
	if !strings.Contains(*query, "where") {
		*query += `where 1 = 1 `
	}
	if sp.Username != nil && *sp.Username != "" {
		*params = append(*params, *sp.Username)
		*query += fmt.Sprintf(" and %v.user_id=$%d", tablePrefix, len(*params))
	}
	if sp.Email != nil && *sp.Email != "" {
		*params = append(*params, *sp.Email)
		*query += fmt.Sprintf(" and %v.email=$%d", tablePrefix, len(*params))
	}
	if sp.UserType != nil && *sp.UserType != "" {
		*params = append(*params, *sp.UserType)
		*query += fmt.Sprintf(" and %v.user_type=$%d", tablePrefix, len(*params))
	}
	if sp.Token != nil && *sp.Token != "" {
		*params = append(*params, *sp.Token)
		*query += fmt.Sprintf(" and %v.token=$%d", tablePrefix, len(*params))
	}
	if !sp.EditInfoCUDSearchParams.IsEmpty() {
		sp.EditInfoCUDSearchParams.appendSearchQuery(tablePrefix, query, params)
	}
}

func (sp *UserSearchParams) appendSortQuery(query *string) {
	if !sp.UserSortParams.IsEmpty() {
		if !strings.Contains(*query, "order by") {
			*query += ` order by `
		}
		*query += sp.UserSortParams.OrderBy()
	}
}

func (sp *UserSearchParams) appendGroupByQuery(query *string) {

}

func (sp *UserSearchParams) appendPagingQuery(query *string, params *[]interface{}) {
	if !sp.PagingSearchParams.IsEmpty() {
		sp.PagingSearchParams.appendSearchQuery(query, params)
	}
}

type UserSortParams struct {
	Prefix   string
	Username *SortColumn `column:"username"`
	UserType *SortColumn `column:"user_type"`
	EditInfoCUDSortParams
}

func (sp UserSortParams) IsEmpty() bool {
	return sp.Username == nil && sp.UserType == nil && sp.EditInfoCUDSortParams.IsEmpty()
}

func (sp UserSortParams) GetTablePrefix() string {
	if sp.Prefix != "" {
		return sp.Prefix
	}
	return "usr"
}

func (sp *UserSortParams) SetTablePrefix(prefix string) {
	sp.Prefix = prefix
}

func (sp UserSortParams) SortColumns() SortColumns {
	var scs SortColumns
	tablePrefix := sp.GetTablePrefix()
	if sp.Username != nil {
		sp.Username.Prefix = tablePrefix
		sp.Username.Column = util.GetTag(sp, "Username", column_tag)
		scs = append(scs, *sp.Username)
	}
	if sp.UserType != nil {
		sp.UserType.Prefix = tablePrefix
		sp.UserType.Column = util.GetTag(sp, "UserType", column_tag)
		scs = append(scs, *sp.UserType)
	}
	if !sp.EditInfoCUDSortParams.IsEmpty() {
		sp.EditInfoCUDSortParams.Prefix = tablePrefix
		scs = append(scs, sp.EditInfoCUDSortParams.SortColumns()...)
	}
	return scs
}

func (sp UserSortParams) OrderBy() string {
	return sp.SortColumns().OrderBy()
}

type UserUpdateParams struct {
	Id                string
	EmailVerified     *int
	PasswordHash      *string
	Token             *string
	TokenValidUntil   *time.Time
	TokenRefreshUntil *time.Time
	EditInfoUDUpdateParams
}

func (up UserUpdateParams) appendUpdateQuery(query *string, params *[]interface{}) {
	*query = `update "user" usr set `

	if up.EmailVerified != nil {
		*params = append(*params, *up.EmailVerified)
		*query += fmt.Sprintf("email_verified = $%d, ", len(*params))
	}

	if up.PasswordHash != nil {
		*params = append(*params, *up.PasswordHash)
		*query += fmt.Sprintf("password_hash = $%d, ", len(*params))
	}

	if up.Token != nil {
		*params = append(*params, up.Token)
		*query += fmt.Sprintf("token = $%d, ", len(*params))
	}

	if up.TokenValidUntil != nil {
		*params = append(*params, up.TokenValidUntil)
		*query += fmt.Sprintf("token_valid_until = $%d, ", len(*params))
	}

	if up.TokenRefreshUntil != nil {
		*params = append(*params, up.TokenRefreshUntil)
		*query += fmt.Sprintf("token_refresh_until = $%d, ", len(*params))
	}

	up.EditInfoUDUpdateParams.appendUpdateQuery(query, params)

	*params = append(*params, up.Id)
	*query += fmt.Sprintf("where usr.user_id = $%d;", len(*params))
}
