package model

func AllModels() []any {
	return []any{
		&UserIdentity{},
		&Tag{},
		&PostTag{},
		&Post{},
		&Comment{},
		&PostAlias{},
		&Like{},
		&SensitiveWord{},
		&ReviewQueue{},
	}
}
