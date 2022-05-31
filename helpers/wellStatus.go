package helpers

import (
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/kmge/paegtm/global"
)

var WellStatusTypeTableName = "dict.well_status_type"

var WellStatusTableName = "prod.well_status"

func WellStatusList(wellId int) (
	data []map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"ws.id",
		"ws.dbeg",
		"ws.dend",
		"wst.name_ru",
		"wst.name_short_ru",
		"wst.code",
	).From(
		fmt.Sprintf("%s ws", WellStatusTableName),
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s wst", WellStatusTypeTableName),
		"ws.status = wst.id",
	).Where(
		sb.Equal("ws.well", wellId),
	).OrderBy("ws.dbeg")

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

func WellStatusGetFirstDate(wellId int) (
	firstDate time.Time,
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.
		Select(
			"min(dbeg)",
		).
		From(fmt.Sprintf("%s ws", WellStatusTableName)).
		Where(sb.Equal("well", wellId))

	query, args := sqlbuilder.
		WithFlavor(sb, sqlbuilder.PostgreSQL).
		Build()

	row := global.Db.QueryRow(query, args...)

	err = row.Scan(&firstDate)

	return
}

func WellStatusGetLastDate(wellId int) (
	lastDate time.Time,
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.
		Select(
			"max(dend)",
		).
		From(fmt.Sprintf("%s ws", WellStatusTableName)).
		Where(
			sb.Equal("well", wellId),
			"dend < date '3333-12-31'",
		)

	query, args := sqlbuilder.
		WithFlavor(sb, sqlbuilder.PostgreSQL).
		Build()

	row := global.Db.QueryRow(query, args...)

	err = row.Scan(&lastDate)

	return
}
