package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/kmge/paegtm/global"
	"github.com/kmge/paegtm/routes/gtm"
	"github.com/kmge/paegtm/routes/measurement"
	"github.com/kmge/paegtm/routes/well"
	"github.com/kmge/paegtm/routes/wellStatus"
)

func main() {
	godotenv.Load()

	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PWD"),
			os.Getenv("POSTGRES_DB"),
			os.Getenv("POSTGRES_PORT"),
		),
	)

	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	} else {
		global.Db = db
	}

	well.AddRoutes()
	gtm.AddRoutes()
	wellStatus.AddRoutes()
	measurement.AddRoutes()

	port, err := strconv.Atoi(os.Getenv("PORT"))

	if err != nil {
		port = 3000
	}

	address := fmt.Sprintf(":%v", port)

	fmt.Printf("Listening %v\n", address)

	log.Fatal(http.ListenAndServe(address, nil))
}
