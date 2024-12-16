package main

import (
	"flag"
	"log"
	"time"

	_ "github.com/lib/pq"

	"importer/config"
	"importer/db"
	"importer/excel"
	"importer/generator"
)

func main() {
	// Parse command line flags
	generateData := flag.Bool("generate", false, "Generate test data")
	numRows := flag.Int("rows", 100000, "Number of rows to generate")
	inputFile := flag.String("file", "test_data.xlsx", "Excel file to process")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if *generateData {
		gen := generator.NewGenerator(generator.GeneratorConfig{
			NumCustomers:    *numRows,
			MultiAcctChance: 0.3,
			ThirdAcctChance: 0.1,
			CustomerPrefix:  "CUST",
			AccountPrefix:   "ACC",
		})

		start := time.Now()
		if err := excel.GenerateFile(*inputFile, gen); err != nil {
			log.Fatal(err)
		}
		log.Printf("Total generation time: %v", time.Since(start))
		return
	}

	// Initialize database
	db, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Process import
	importer := excel.NewImporter(db, cfg)
	if err := importer.Import(*inputFile); err != nil {
		log.Fatal(err)
	}
}