package share

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

//	type UserContext struct {
//		UserID               float64
//		UserUuid             string
//		UserName             string
//		RoleId               uint64
//		LoginSession         string
//		UserAvatar           string
//		Exp                  time.Time
//		KeyAliasForWebsocket string
//		UserAgent            string
//		Ip                   string
//	}
type MemberSession struct {
	MemberID     int64  `db:"id" json:"member_id"`
	MemberUUID   string `db:"member_uuid" json:"member_uuid"`
	LoginID      string `db:"user_name" json:"login_id"`
	Avatar       string `db:"user_avatar" json:"avatar"`
	RoleID       int64  `db:"role_id" json:"role_id"`
	LoginSession string `db:"login_session" json:"login_session"`
}
type MemberContext struct {
	MemberID             float64   `json:"member_id"`
	MemberUUID           string    `json:"member_uuid"`
	LoginID              string    `json:"login_id"`
	Avatar               string    `json:"avatar"`
	RoleID               uint64    `json:"role_id"`
	LoginSession         string    `json:"login_session"`
	KeyAliasForWebsocket string    `json:"key_alias"`
	Exp                  time.Time `json:"exp"`
	UserAgent            string    `json:"user_agent"`
	IP                   string    `json:"ip"`
}

type Paging struct {
	Page    int `json:"page" query:"page" validate:"required,min=1"`
	Perpage int `json:"per_page" query:"per_page" validate:"required,min=1"`
}

type Sort struct {
	Property  string `json:"property" validate:"required"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"`
}
type Filter struct {
	Property string      `json:"property" validate:"required"`
	Value    interface{} `json:"value" validate:"required"`
}

type FieldId struct {
	Id uint64 `json:"id"`
}

type FieldFunctionIds struct {
	FunctionIDs string `json:"function_ids"`
}

type Status struct {
	Id         int    `json:"id"`
	StatusName string `json:"status_name"`
}

type BroadcastResponse struct {
	Topic string          `json:"topic"`
	Data  json.RawMessage `json:"data"`
}

var StatusData = []Status{
	{Id: 1, StatusName: "Active"},
	{Id: 2, StatusName: "Inactive"},
	{Id: 3, StatusName: "Suspended"},
	{Id: 4, StatusName: "Deleted"},
}

// Platform Mini
type Platform struct {
	ID                     uint64    `json:"id"`
	MembershipPlatformUUID uuid.UUID `json:"membership_platform_uuid"`
	PlatformName           string    `json:"platform_name"`
	PlatformHost           string    `json:"platform_host"`
	PlatformToken          string    `json:"platform_token"`
	PlatformExtraPayload   string    `json:"platform_extra_payload"`
	InternalToken          string    `json:"internal_token"`
	StatusID               uint64    `json:"status_id"`
	Order                  uint64    `json:"order"`
}
