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
	default:
		fmt.Println("invalid argument")
		fmt.Println("usage: [--new-nonce]")
		os.Exit(0)
	}
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
			Name:   result[0],
			Abbr:   result[1],
			Lat:    float32(lat),
			Lon:    float32(lon),
			Domain: result[4],
		}

		schools = append(schools, school)
	}

	pg := db.New()
	fmt.Print("Delete existing school table")
	pg.Exec("DELETE FROM schools")
	fmt.Println("\tOK")

	fmt.Print("Adding data")
	result := pg.Model(&[]db.School{}).Create(schools)
	if result.Error != nil {
		fmt.Println("\t\t\tERROR")
		panic(result.Error.Error())
	}
	fmt.Println("\t\t\tOK")
}

func nonceGenerator() {
	nonce := make([]byte, 12)

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	nonceStr := hex.EncodeToString(nonce)
	fmt.Println(nonceStr)
}
