package helpers

import (
	"fmt"
	"math"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/kmge/paegtm/global"
)

var MeasLiqTableName = "prod.meas_liq"
var MeasWaterCutTableName = "prod.meas_water_cut"

func MeasurementList(wellId int) (
	data []map[string]interface{},
	err error,
) {
	startDate, err := WellStatusGetFirstDate(wellId)

	if err != nil {
		return
	} else {
		startDate = time.Date(
			startDate.Year(),
			startDate.Month(),
			1,
			0, 0, 0, 0,
			startDate.Location(),
		)
	}

	gtm, err := GTMGetLast(wellId)

	if err != nil {
		return
	}

	endDate := startDate.AddDate(1, 0, 0)

	if gtm != nil {
		date, ok := gtm["dend"].(time.Time)

		if ok {
			if date.Day() == 1 {
				endDate = date.AddDate(1, -1, 0)
			} else {
				endDate = date.AddDate(1, 0, 0)
			}
		}
	}

	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"dd as date",
		`
			sum(
				distinct
				extract(
					epoch from (
						least(ws.dend, dd + interval '1 month') 
						- 
						greatest(ws.dbeg, dd)
					)
				)
			)/3600/24 as work
		`,
		`
			sum(
				coalesce(
					ml.liquid
					*
					extract(
						epoch from (
							least(
								coalesce(ml.dend, mwc.dend, dd), 	
								coalesce(mwc.dend, ml.dend, dd),							
								ws.dend, 
								dd + interval '1 month'
							) 
							- 
							greatest(
								coalesce(ml.dbeg, mwc.dbeg, dd), 
								coalesce(mwc.dbeg, ml.dbeg, dd),
								ws.dbeg, 
								dd
							)
						)
					)
					/
					3600
					/
					24
				, 0)
			) as liquid		
		`,
		`
			(sum(
				coalesce(
					ml.liquid 
					* 
					coalesce(1 - mwc.water_cut/100, 1) 
					*
					extract(
						epoch from (
							least(
								coalesce(ml.dend, mwc.dend, dd), 	
								coalesce(mwc.dend, ml.dend, dd),							
								ws.dend,
								dd + interval '1 month'
							) 
							- 
							greatest(
								coalesce(ml.dbeg, mwc.dbeg, dd), 
								coalesce(mwc.dbeg, ml.dbeg, dd),
								ws.dbeg, 
								dd
							)
						)
					)
					/
					3600
					/
					24
				, 0)
			)) as oil		
		`,
	).From(
		fmt.Sprintf(
			"generate_series('%s'::timestamp, '%s'::timestamp, '1 month'::interval) dd",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"),
		),
	).Join(
		fmt.Sprintf("%s wst", WellStatusTypeTableName),
		"wst.code = 'WRK'",
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s ws", WellStatusTableName),
		sb.Equal("ws.well", wellId),
		"ws.dend >= ws.dbeg",
		"ws.dbeg <= dd + interval '1 month'",
		"ws.dend >= dd",
		"ws.dend < date '3333-12-31'",
		"ws.status = wst.id",
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s ml", MeasLiqTableName),
		sb.Equal("ml.well", wellId),
		"ml.dend >= ml.dbeg",
		"ml.dbeg <= ws.dend",
		"ml.dend >= ws.dbeg",
		"ml.dbeg <= dd + interval '1 month'",
		"ml.dend >= dd",
		"ml.dend < date '3333-12-31'",
	).JoinWithOption(
		sqlbuilder.LeftOuterJoin,
		fmt.Sprintf("%s mwc", MeasWaterCutTableName),
		sb.Equal("mwc.well", wellId),
		"mwc.dend >= mwc.dbeg",
		"mwc.dbeg <= ml.dend",
		"mwc.dend >= ml.dbeg",
		"mwc.dbeg <= ws.dend",
		"mwc.dend >= ws.dbeg",
		"mwc.dbeg <= dd + interval '1 month'",
		"mwc.dend >= dd",
		"mwc.dend < date '3333-12-31'",
	).GroupBy(
		"dd",
	).OrderBy(
		"dd",
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

func MeasurementLiquidLast(wellId int) (
	liqMeas map[string]interface{},
	err error,
) {
	sb := sqlbuilder.NewSelectBuilder()

	sb.Select(
		"id", "dbeg", "dend", "liquid",
	).From(
		MeasLiqTableName,
	).Where(
		sb.Equal("well", wellId),
	).OrderBy(
		"dbeg desc",
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
		liqMeas = data[0]
	}

	return
}

func MeasurementProcessData(
	well map[string]interface{},
	gtm map[string]interface{},
	measurementList []map[string]interface{},
	lastDate time.Time,
	hasLastDate bool,
	liquidRateBeforeGTM float64,
	oilRateBeforeGTM float64,
) []map[string]interface{} {
	gtmYearMonth := DateMonthYearFormat(gtm["dbeg"])
	lastDateYearMonth := DateMonthYearFormat(lastDate)

	result := make([]map[string]interface{}, len(measurementList))

	totalWorkBase := 0.
	totalWorkPred := 0.

	avgIncProdLiquidPred := 0.
	avgIncProdOilPred := 0.

	for i, item := range measurementList {
		work := item["work"].(float64)

		result[i] = make(map[string]interface{})
		result[i]["date"] = item["date"]
		result[i]["work"] = item["work"]

		liquid := item["liquid"].(float64)
		oil := item["oil"].(float64)
		baseLiquid := liquid
		baseOil := oil

		result[i]["fact_prediction"] = "Факт"

		if hasLastDate &&
			(DateMonthYearFormat(item["date"]) == DateMonthYearFormat(lastDate) ||
				item["date"].(time.Time).After(lastDate)) {

			result[i]["fact_prediction"] = "Прогноз"
		}

		if gtm != nil && gtm["dend"] != nil {
			if item["date"].(time.Time).After(gtm["dend"].(time.Time)) {
				totalWorkBase += work

				baseLiquid = 0.
				baseOil = 0.

				if totalWorkBase > 0 {
					baseLiquid = liquidRateBeforeGTM * (1 - math.Exp(-0.000276509*totalWorkBase)) * work / 0.000276509 / totalWorkBase
					baseOil = oilRateBeforeGTM * (1 - math.Exp(-0.000276509*totalWorkBase)) * work / 0.000276509 / totalWorkBase
				}
			}
		}

		addProdLiquid := liquid - baseLiquid
		addProdOil := oil - baseOil

		avgIncProdLiquid := 0.
		avgIncProdOil := 0.

		if work > 0 {
			avgIncProdLiquid = addProdLiquid / work
			avgIncProdOil = addProdOil / work
		}

		if i > 0 && hasLastDate && (DateMonthYearFormat(item["date"]) == DateMonthYearFormat(lastDate) ||
			item["date"].(time.Time).After(lastDate)) {

			totalWorkPred += work
			avgIncProdLiquid = avgIncProdLiquidPred * (1 - math.Exp(-0.000276509*totalWorkPred)) / 0.000276509 / totalWorkPred
			avgIncProdOil = avgIncProdOilPred * (1 - math.Exp(-0.000276509*totalWorkPred)) / 0.000276509 / totalWorkPred

			addProdLiquid = avgIncProdLiquid * work
			addProdOil = avgIncProdOil * work

			liquid = baseLiquid + addProdLiquid
			oil = baseOil + addProdOil
		} else {
			avgIncProdLiquidPred = avgIncProdLiquid
			avgIncProdOilPred = avgIncProdOil
		}

		result[i]["liquid"] = liquid
		result[i]["oil"] = oil
		result[i]["base_liquid"] = baseLiquid
		result[i]["base_oil"] = baseOil

		result[i]["add_prod_liquid"] = addProdLiquid
		result[i]["add_prod_oil"] = addProdOil

		result[i]["avg_inc_prod_liquid"] = avgIncProdLiquid
		result[i]["avg_inc_prod_oil"] = avgIncProdOil

		result[i]["total_work_base"] = totalWorkBase

		itemYearMonth := DateMonthYearFormat(item["date"])

		result[i]["is_gtm"] = itemYearMonth == gtmYearMonth

		if hasLastDate && lastDateYearMonth == itemYearMonth {
			result[i]["is_last"] = true
		} else {
			result[i]["is_last"] = false
		}
	}

	return result
}

func MeasurementBeforeGTM(
	wellId int,
	gtm map[string]interface{},
) (
	work float64,
	liquid float64,
	oil float64,
	err error,
) {
	if gtm != nil && gtm["dbeg"] != nil {
		sb := sqlbuilder.NewSelectBuilder()

		sb.Select(
			fmt.Sprintf(
				`
					sum(
						distinct
						extract(
							epoch from (
								least(ws.dend, date '%s') 
								- 
								greatest(ws.dbeg, date '%s'  - interval '91 day')
							)
						)
					)/3600/24 as work
				`,
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
			),
			fmt.Sprintf(
				`
					sum(
						coalesce(
							ml.liquid
							*
							extract(
								epoch from (
									least(
										coalesce(ml.dend, mwc.dend, date '%s' - interval '91 day'), 	
										coalesce(mwc.dend, ml.dend, date '%s' - interval '91 day'),							
										ws.dend, 
										date '%s'
									) 
									- 
									greatest(
										coalesce(ml.dbeg, mwc.dbeg, date '%s' - interval '91 day'), 
										coalesce(mwc.dbeg, ml.dbeg, date '%s' - interval '91 day'),
										ws.dbeg, 
										date '%s'  - interval '91 day'
									)
								)
							)
							/
							3600
							/
							24
						, 0)
					) as liquid		
				`,
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
			),
			fmt.Sprintf(
				`
					(sum(
						coalesce(
							ml.liquid 
							* 
							coalesce(1 - mwc.water_cut/100, 1) 
							*
							extract(
								epoch from (
									least(
										coalesce(ml.dend, mwc.dend, date '%s'  - interval '91 day'), 	
										coalesce(mwc.dend, ml.dend, date '%s' - interval '91 day'),							
										ws.dend,
										date '%s'
									) 
									- 
									greatest(
										coalesce(ml.dbeg, mwc.dbeg, date '%s' - interval '91 day'), 
										coalesce(mwc.dbeg, ml.dbeg, date '%s' - interval '91 day'),
										ws.dbeg, 
										date '%s'  - interval '91 day'
									)
								)
							)
							/
							3600
							/
							24
						, 0)
					)) as oil		
				`,
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
				gtm["dbeg"].(time.Time).Format("2006-01-02"),
			),
		).
			From(
				fmt.Sprintf("%s ws", WellStatusTableName),
			).
			Join(
				fmt.Sprintf("%s wst", WellStatusTypeTableName),
				"wst.code = 'WRK'",
				"ws.status = wst.id",
			).
			JoinWithOption(
				sqlbuilder.LeftOuterJoin,
				fmt.Sprintf("%s ml", MeasLiqTableName),
				sb.Equal("ml.well", wellId),
				"ml.dend >= ml.dbeg",
				"ml.dbeg <= ws.dend",
				"ml.dend >= ws.dbeg",
				fmt.Sprintf("ml.dbeg <= date '%s'", gtm["dbeg"].(time.Time).Format("2006-01-02")),
				fmt.Sprintf("ml.dend >= date '%s' - interval '91 day'", gtm["dbeg"].(time.Time).Format("2006-01-02")),
				"ml.dend < date '3333-12-31'",
			).JoinWithOption(
			sqlbuilder.LeftOuterJoin,
			fmt.Sprintf("%s mwc", MeasWaterCutTableName),
			sb.Equal("mwc.well", wellId),
			"mwc.dend >= mwc.dbeg",
			"mwc.dbeg <= ml.dend",
			"mwc.dend >= ml.dbeg",
			"mwc.dbeg <= ws.dend",
			"mwc.dend >= ws.dbeg",
			fmt.Sprintf("mwc.dbeg <= date '%s'", gtm["dbeg"].(time.Time).Format("2006-01-02")),
			fmt.Sprintf("mwc.dend >= date '%s' - interval '91 day'", gtm["dbeg"].(time.Time).Format("2006-01-02")),
			"mwc.dend < date '3333-12-31'",
		).
			Where(
				sb.Equal("ws.well", wellId),
				"ws.dend >= ws.dbeg",
				fmt.Sprintf("ws.dbeg <= date '%s'", gtm["dbeg"].(time.Time).Format("2006-01-02")),
				fmt.Sprintf("ws.dend >= date '%s' - interval '91 day'", gtm["dbeg"].(time.Time).Format("2006-01-02")),
				"ws.dend < date '3333-12-31'",
			)

		query, args := sqlbuilder.
			WithFlavor(sb, sqlbuilder.PostgreSQL).
			Build()

		row := global.Db.QueryRow(query, args...)

		err = row.Scan(&work, &liquid, &oil)
	}

	return
}
