package wellStatus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kmge/paegtm/helpers"
)

func List(w http.ResponseWriter, r *http.Request) {
	wellIdStr := r.URL.Path[len("/well-status/list/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		var wellStatusList []map[string]interface{}

		wellStatusList, err = helpers.WellStatusList(wellId)

		if err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			data, _ := json.Marshal(wellStatusList)

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

func FirstDate(w http.ResponseWriter, r *http.Request) {
	wellIdStr := r.URL.Path[len("/well-status/first-date/"):]
	wellId, err := strconv.Atoi(wellIdStr)

	if err == nil {
		var firstDate time.Time

		firstDate, err = helpers.WellStatusGetFirstDate(wellId)

		if err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")

			data, _ := json.Marshal(firstDate)

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

func AddRoutes() {
	http.HandleFunc("/well-status/list/", List)
	http.HandleFunc("/well-status/first-date/", FirstDate)
}
