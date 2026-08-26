package main

import (
	"flag"
	"log"
	"net/http"

	"strata-weave/internal/handler"
	"strata-weave/internal/service"
	"strata-weave/internal/store"
)

func main() {
	dbPath := flag.String("db", "strata.db", "SQLite database path")
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := service.New(db)
	mux := handler.New(app)
	log.Printf("strata-weave listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
