package main

import (
	"bytes"
	"confesi/db"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm/clause"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing arg")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--new-nonce":
		nonceGenerator()
	case "--seed-schools":
		seed_schools()
	case "--seed-all":
		seed_all()
	case "--seed-feedback-types":
		seed_feedback_types()
	case "--seed-report-types":
		seed_report_types()
	case "--seed-post-categories":
		seed_post_categories()
	case "--help":
		fmt.Println("usage: [--new-nonce] [--seed-schools] [--seed-all] [--seed-feedback-types] [--seed-report-types] [--seed-post-categories]")
	default:
		fmt.Println("invalid argument")
		fmt.Println("usage: [--help]")
		os.Exit(0)
	}
}

func seed_all() {
	seed_schools()
	seed_feedback_types()
	seed_report_types()
	seed_post_categories()
}

func seed_post_categories() {
	f, err := os.ReadFile("./seeds/seed_post_categories.csv")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	categories := []db.PostCategory{}
	reader.Read() // skips header

	for {

		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(nil)
		}

		category := db.PostCategory{
			Name: result[0],
		}

		categories = append(categories, category)
	}

	pg := db.New()

	result := pg.
		Model(&[]db.PostCategory{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(categories)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Post Categories Completed")
	fmt.Printf(fmt.Sprintf("Rows Updates: %v / %v", result.RowsAffected, len(categories)))
}

func seed_report_types() {
	f, err := os.ReadFile("./seeds/seed_report_types.csv")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	types := []db.ReportType{}
	reader.Read() // skips header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(nil)
		}

		report_type := db.ReportType{
			Type: result[0],
		}

		types = append(types, report_type)
	}

	pg := db.New()

	result := pg.
		Model(&[]db.ReportType{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(types)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Report Types Completed")
	fmt.Printf(fmt.Sprintf("Rows Updates: %v / %v", result.RowsAffected, len(types)))
}

func seed_feedback_types() {
	f, err := os.ReadFile("./seeds/seed_feedback_types.csv")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	types := []db.FeedbackType{}
	reader.Read() // skips header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(nil)
		}

		feedback_type := db.FeedbackType{
			Type: result[0],
		}

		types = append(types, feedback_type)
	}

	pg := db.New()

	result := pg.
		Model(&[]db.FeedbackType{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(types)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Feedback Types Completed")
	fmt.Printf(fmt.Sprintf("Rows Updates: %v / %v", result.RowsAffected, len(types)))
}

func seed_schools() {
	f, err := os.ReadFile("./seeds/seed_schools.csv")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	schools := []db.School{}
	reader.Read() // skips header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(nil)
		}

		lat, err := strconv.ParseFloat(result[2], 32)
		if err != nil {
			panic(err)
		}

		lon, err := strconv.ParseFloat(result[3], 32)
		if err != nil {
			panic(err)
		}

		school := db.School{
			Name:     result[0],
			Abbr:     result[1],
			Lat:      float32(lat),
			Lon:      float32(lon),
			Domain:   result[4],
			ImgUrl:   result[5],
			Website:  result[6],
			Timezone: result[7],
		}

		schools = append(schools, school)
	}

	pg := db.New()

	// Create the schools rows in the schools table, but on conflict do not create

	result := pg.
		Model(&[]db.School{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(schools)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Schools Completed")
	fmt.Printf(fmt.Sprintf("Rows Updates: %v / %v", result.RowsAffected, len(schools)))
}

func nonceGenerator() {
	nonce := make([]byte, 12)

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	nonceStr := hex.EncodeToString(nonce)
	fmt.Println(nonceStr)
}
