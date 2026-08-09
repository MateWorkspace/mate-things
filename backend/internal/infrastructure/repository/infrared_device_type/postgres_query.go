package infrastructurerepositoryinfrareddevicetype

import "github.com/Masterminds/squirrel"

func (p *postgresImpl) queryList() (query string, args []any, err error) {
	return p.SqrD.Select("id", "name").From("infrared_device_type").OrderBy("name").ToSql()
}

func (p *postgresImpl) queryReadByName(name string) (query string, args []any, err error) {
	return p.SqrD.Select("id", "name").From("infrared_device_type").Where(squirrel.Eq{"name": name}).ToSql()
}

func (p *postgresImpl) queryCreate(name string) (query string, args []any, err error) {
	return p.SqrD.Insert("infrared_device_type").Columns("name").Values(name).Suffix("RETURNING id").ToSql()
}
