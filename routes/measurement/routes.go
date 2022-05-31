package measurement

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/kmge/paegtm/helpers"
)

var templates = template.Must(
	template.
		New("view").
		Funcs(
			template.FuncMap{
				"DateFormat":          helpers.DateFormat,
				"DateMonthYearFormat": helpers.DateMonthYearFormat,
				"DateEq":              helpers.DateEq,
				"NumTruncate":         helpers.NumTruncate,
			},
		).
		ParseFiles(
			"routes/measurement/view.html",
		),
)

func List(w http.ResponseWriter, r *http.Request) {
	wellIdStr := r.URL.Path[len("/measurement/list/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		var measurementList []map[string]interface{}

		measurementList, err = helpers.MeasurementList(wellId)

		if err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			data, _ := json.Marshal(measurementList)

			fmt.Fprint(w, string(data))

			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	http.Error(
		w,
		err.Error(),
		http.StatusBadRequest,
	)
}

func LiquidLast(w http.ResponseWriter, r *http.Request) {
	var liqMeas map[string]interface{}

	wellIdStr := r.URL.Path[len("/measurement/liquid/last/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		liqMeas, err = helpers.MeasurementLiquidLast(wellId)

		if liqMeas != nil && err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			data, _ := json.Marshal(liqMeas)

			fmt.Fprint(w, string(data))

			return
		}
	}

	errorTxt := fmt.Sprintf("Замеры добычи жидкости для wellId=%v не существет", wellId)

	if err != nil {
		errorTxt = err.Error()
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	http.Error(
		w,
		errorTxt,
		http.StatusBadRequest,
	)
}

func View(w http.ResponseWriter, r *http.Request) {
	wellIdStr := r.URL.Path[len("/measurement/view/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		var measurementList []map[string]interface{}

		measurementList, err = helpers.MeasurementList(wellId)

		if err == nil {
			var well map[string]interface{}

			well, err = helpers.WellGetById(wellId)

			if err == nil {
				var gtm map[string]interface{}

				gtm, err = helpers.GTMGetLast(wellId)

				if err == nil {
					var workBeforeGTM, liquidBeforeGTM, oilBeforeGTM float64

					workBeforeGTM, liquidBeforeGTM, oilBeforeGTM, err = helpers.MeasurementBeforeGTM(wellId, gtm)

					liquidRateBeforeGTM := 0.
					oilRateBeforeGTM := 0.

					if workBeforeGTM > 0 {
						liquidRateBeforeGTM = liquidBeforeGTM / workBeforeGTM
						oilRateBeforeGTM = oilBeforeGTM / workBeforeGTM
					}

					if err == nil {
						lastDate, errLastDate := helpers.WellStatusGetLastDate(wellId)

						data := helpers.MeasurementProcessData(
							well,
							gtm,
							measurementList,
							lastDate,
							errLastDate == nil,
							liquidRateBeforeGTM,
							oilRateBeforeGTM,
						)

						err = templates.ExecuteTemplate(
							w,
							"view.html",
							map[string]interface{}{
								"data":        data,
								"well":        well,
								"gtm":         gtm,
								"lastDate":    lastDate,
								"hasLastDate": errLastDate == nil,
							},
						)

						if err == nil {
							return
						}
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	http.Error(
		w,
		err.Error(),
		http.StatusBadRequest,
	)
}

func AddRoutes() {
	http.HandleFunc("/measurement/view/", View)
	http.HandleFunc("/measurement/list/", List)
	http.HandleFunc("/measurement/liquid/last/", LiquidLast)
}
