package gtm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kmge/paegtm/helpers"
)

func List(w http.ResponseWriter, r *http.Request) {
	wellIdStr := r.URL.Path[len("/gtm/list/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		var gtmList []map[string]interface{}

		gtmList, err = helpers.GTMList(wellId)

		if err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			data, _ := json.Marshal(gtmList)

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

func Last(w http.ResponseWriter, r *http.Request) {
	var gtm map[string]interface{}

	wellIdStr := r.URL.Path[len("/gtm/last/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		gtm, err = helpers.GTMGetLast(wellId)

		if gtm != nil && err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			data, _ := json.Marshal(gtm)

			fmt.Fprint(w, string(data))

			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	errorTxt := fmt.Sprintf("ГТМ для wellId=%v не существет", wellId)

	if err != nil {
		errorTxt = err.Error()
	}

	http.Error(
		w,
		errorTxt,
		http.StatusBadRequest,
	)
}

func AddRoutes() {
	http.HandleFunc("/gtm/list/", List)
	http.HandleFunc("/gtm/last/", Last)
}
