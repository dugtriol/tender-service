package service

import "fmt"

var (
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
	ErrCannotCreateUser  = fmt.Errorf("cannot create user")
	ErrUserNotFound      = fmt.Errorf("user not found")
	ErrCannotGetUser     = fmt.Errorf("cannot get user")

	ErrOrgAlreadyExists = fmt.Errorf("organization already exists")
	ErrCannotCreateOrg  = fmt.Errorf("cannot create organization")
	ErrOrgNotFound      = fmt.Errorf("organization not found")
	ErrCannotGetOrg     = fmt.Errorf("cannot get organization")

	ErrOrgRespAlreadyExists = fmt.Errorf("organization responsible already exists")
	ErrCannotCreateOrgResp  = fmt.Errorf("cannot create organization responsible")
	ErrOrgRespNotFound      = fmt.Errorf("organization responsible not found")
	ErrCannotGetOrgResp     = fmt.Errorf("cannot get organization responsible")

	ErrTenderAlreadyExists = fmt.Errorf("tender already exists")
	ErrCannotCreateTender  = fmt.Errorf("cannot create tender")
	ErrTenderNotFound      = fmt.Errorf("tender not found")
	ErrCannotGetTender     = fmt.Errorf("cannot get tender")
	ErrCannotPutStatus     = fmt.Errorf("cannot put status")
	ErrCannotEditTender    = fmt.Errorf("cannot edit tender")
	ErrCannotIncrement     = fmt.Errorf("cannot incremet")

	ErrBidAlreadyExists = fmt.Errorf("tender already exists")
	ErrCannotCreateBid  = fmt.Errorf("cannot create tender")
	ErrBidNotFound      = fmt.Errorf("tender not found")
	ErrCannotGetBid     = fmt.Errorf("cannot get tender")
	ErrCannotEditBid    = fmt.Errorf("cannot edit tender")
)
