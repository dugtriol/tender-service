package entity

type OrgResponsible struct {
	Id             string `db:"id"`
	OrganizationId string `db:"organization_id"`
	UserId         string `db:"user_id"`
}
