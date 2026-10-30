package main

import (
	"flag"
	"log"

	"github.com/djsilvajr/simple-project-template/back/internal/config"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/database"
	"github.com/djsilvajr/simple-project-template/back/migration"
)

// Uso:
//
//	migration              aplica as migrations
//	migration -seed        aplica as migrations e executa os seeders
//	migration -seed-only   executa apenas os seeders
func main() {
	seed := flag.Bool("seed", false, "executa os seeders após as migrations")
	seedOnly := flag.Bool("seed-only", false, "executa apenas os seeders")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	if !*seedOnly {
		if err := migration.Migrate(db); err != nil {
			log.Fatalf("migration: %v", err)
		}
		log.Println("migration: concluída")
	}

	if *seed || *seedOnly {
		if err := migration.Seed(db, cfg); err != nil {
			log.Fatalf("seeder: %v", err)
		}
		log.Println("seeder: concluído")
	}
}
