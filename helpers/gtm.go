package helpers

import (
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/kmge/paegtm/global"
)

var GTMTypeTableName = "dict.gtm_type"
var GTMKindTableName = "dict.gtm_kind"

var GTMTableName = "prod.gtm"

func GTMList(wellId int) (
	data []map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.
		Select(
			"g.id",
			"gt.name_ru as type_name_ru",
			"gt.name_short_ru as type_name_short_ru",
			"gk.name_ru as kind_name_ru",
			"gk.name_short_ru as kind_name_short_ru",
			"gk.code as kind_code",
			"g.dbeg",
			"g.dend",
		).
		From(fmt.Sprintf("%s g", GTMTableName)).
		JoinWithOption(
			sqlbuilder.LeftOuterJoin,
			fmt.Sprintf("%s gt", GTMTypeTableName),
			"g.gtm_type = gt.id",
		).
		JoinWithOption(
			sqlbuilder.LeftOuterJoin,
			fmt.Sprintf("%s gk", GTMKindTableName),
			"gt.gtm_kind = gk.id",
		).
		Where(sb.Equal("g.well", wellId)).
		OrderBy("g.dbeg")

	query, args := sqlbuilder.
		WithFlavor(sb, sqlbuilder.PostgreSQL).
		Build()

	rows, err := global.Db.Query(query, args...)

	if err != nil {
		return
	}

	defer rows.Close()

	data, err = DBGetDataFromRows(rows)

	return
}

func GTMGetLast(wellId int) (
	gtm map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"g.id",
		"gt.name_ru as type_name_ru",
		"gt.name_short_ru as type_name_short_ru",
		"gk.name_ru as kind_name_ru",
		"gk.name_short_ru as kind_name_short_ru",
		"gk.code as kind_code",
		"g.dbeg",
		"g.dend",
	).From(
		fmt.Sprintf("%s g", GTMTableName),
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s gt", GTMTypeTableName),
		"g.gtm_type = gt.id",
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s gk", GTMKindTableName),
		"gt.gtm_kind = gk.id",
	).Where(
		sb.Equal("g.well", wellId),
	).OrderBy(
		"g.dbeg desc",
	).Limit(1)

	query, args := sqlbuilder.
		WithFlavor(sb, sqlbuilder.PostgreSQL).
		Build()

	rows, err := global.Db.Query(query, args...)

	if err != nil {
		return
	}

	defer rows.Close()

	data, err := DBGetDataFromRows(rows)

	if len(data) > 0 {
		gtm = data[0]
	}

	return
}
