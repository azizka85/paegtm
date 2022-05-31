package helpers

import (
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"github.com/kmge/paegtm/global"
)

var WellTableName = "dict.well"
var WellTypeTableName = "dict.well_type"

var WellListCountPerPage = 100

func WellList(pageIndex int) (
	data []map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"w.id", "w.uwi", "w.project_date",
		"wt.name_ru as well_type_name_ru",
		"wt.name_short_ru as well_type_name_short_ru",
	).From(
		fmt.Sprintf("%s w", WellTableName),
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s wt", WellTypeTableName),
		"wt.id = w.well_type",
	).OrderBy(
		"w.uwi",
	).Offset(
		(pageIndex - 1) * WellListCountPerPage,
	).Limit(
		WellListCountPerPage,
	)

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

func WellListCount() (
	count int,
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"count(*)",
	).From(
		WellTableName,
	)

	query, args := sqlbuilder.
		WithFlavor(sb, sqlbuilder.PostgreSQL).
		Build()

	row := global.Db.QueryRow(query, args...)

	err = row.Scan(&count)

	return
}

func WellGet(uwi string) (
	well map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"id", "uwi", "project_date",
	).From(
		WellTableName,
	).Where(
		sb.Equal("uwi", uwi),
	)

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
		well = data[0]
	}

	return
}

func WellGetById(id int) (
	well map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"uwi", "project_date",
	).From(
		WellTableName,
	).Where(
		sb.Equal("id", id),
	)

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
		well = data[0]
	}

	return
}
