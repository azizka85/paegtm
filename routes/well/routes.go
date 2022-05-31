package well

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
		New("list").
		Funcs(
			template.FuncMap{
				"DateFormat": helpers.DateFormat,
			},
		).
		ParseFiles(
			"routes/well/list.html",
		),
)

func Default(w http.ResponseWriter, r *http.Request) {
	uwi := r.URL.Path[len("/well/"):]

	well, err := helpers.WellGet(uwi)

	if well == nil || err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		errorTxt := fmt.Sprintf("Скважина для uwi=%s не существет", uwi)

		if err != nil {
			errorTxt = err.Error()
		}

		http.Error(
			w,
			errorTxt,
			http.StatusBadRequest,
		)
	} else {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")

		data, _ := json.Marshal(well)

		fmt.Fprint(w, string(data))
	}
}

func ViewList(w http.ResponseWriter, r *http.Request) {
	pageIndex, _ := strconv.Atoi(r.URL.Path[len("/well/view/list/"):])

	if pageIndex == 0 {
		pageIndex = 1
	}

	wellList, err := helpers.WellList(pageIndex)
	count, _ := helpers.WellListCount()

	if err == nil {
		err = templates.ExecuteTemplate(
			w,
			"list.html",
			map[string]interface{}{
				"data":  wellList,
				"pages": helpers.RangeMake(1, count/helpers.WellListCountPerPage),
			},
		)

		if err == nil {
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

func AddRoutes() {
	http.HandleFunc("/well/view/list/", ViewList)
	http.HandleFunc("/well/", Default)
}
