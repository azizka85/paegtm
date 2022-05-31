package helpers

import "database/sql"

func DBGetDataFromRows(rows *sql.Rows) (
	data []map[string]interface{},
	err error,
) {
	cols, err := rows.Columns()

	if err != nil {
		return
	}

	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))

	for i, _ := range columns {
		columnPointers[i] = &columns[i]
	}

	for rows.Next() {
		if err = rows.Scan(columnPointers...); err != nil {
			return
		}

		row := make(map[string]interface{})

		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			row[colName] = *val
		}

		data = append(data, row)
	}

	return data, err
}
