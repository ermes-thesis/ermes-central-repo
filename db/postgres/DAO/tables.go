package DAO

type User struct {
	Id      int    `db:"id"`
	Name    string `db:"name"`
	Surname string `db:"surname"`
}

type Group struct {
	GroupName string `db:"group_name"`
}

type UserGroupRelation struct {
	UserId        int    `db:"user_id"`
	AssignedGroup string `db:"assigned_group"`
}
