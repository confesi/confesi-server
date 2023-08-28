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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/guregu/null.v4"
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
	case "--seed-faculties":
		seedFaculties()
	case "--seed-schools":
		seedSchools()
	case "--test-endpoints-speed":
		testEndpointsSpeed()
	case "--seed-all":
		seedAll()
	case "--seed-feedback-types":
		seedFeedbackTypes()
	case "--seed-report-types":
		seedReportTypes()
	case "--seed-post-categories":
		seedPostCategories()
	case "--seed-years-of-study":
		seedYearOfStudies()
	case "--help":
		fmt.Println("usage: [--new-nonce] [--seed-schools] [--seed-all] [--seed-feedback-types] [--seed-report-types] [--seed-post-categories]")
	default:
		fmt.Println("invalid argument")
		fmt.Println("usage: [--help]")
		os.Exit(0)
	}
}

func seedAll() {
	seedSchools()
	seedFeedbackTypes()
	seedReportTypes()
	seedPostCategories()
	seedFaculties()
	seedYearOfStudies()
}

func seedYearOfStudies() {
	f, err := os.ReadFile("./seeds/seed_year_of_study.csv")
	if err != nil {
		log.Fatal(err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	yearOfStudies := []db.YearOfStudy{}
	reader.Read() // skips header

	for {

		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(nil)
		}

		yearOfStudy := db.YearOfStudy{
			Name: null.NewString(result[0], true),
		}

		yearOfStudies = append(yearOfStudies, yearOfStudy)
	}

	pg := db.New()

	result := pg.
		Model(&[]db.YearOfStudy{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(yearOfStudies)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Years of Study Completed")
	fmt.Printf(fmt.Sprintf("Rows Updates: %v / %v", result.RowsAffected, len(yearOfStudies)))
}

func seedFaculties() {
	f, err := os.ReadFile("./seeds/seed_faculties.csv")
	if err != nil {
		log.Fatal(err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	faculties := []db.Faculty{}
	reader.Read() // skips header

	for {

		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(nil)
		}

		faculty := db.Faculty{
			Faculty: null.NewString(result[0], true),
		}

		faculties = append(faculties, faculty)
	}

	pg := db.New()

	result := pg.
		Model(&[]db.Faculty{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(faculties)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Faculties Completed")
	fmt.Printf(fmt.Sprintf("Rows Updates: %v / %v", result.RowsAffected, len(faculties)))
}

func seedPostCategories() {
	f, err := os.ReadFile("./seeds/seed_post_categories.csv")
	if err != nil {
		log.Fatal(err)
		return
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

func seedReportTypes() {
	f, err := os.ReadFile("./seeds/seed_report_types.csv")
	if err != nil {
		log.Fatal(err)
		return
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

func seedFeedbackTypes() {
	f, err := os.ReadFile("./seeds/seed_feedback_types.csv")
	if err != nil {
		log.Fatal(err)
		return
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

func seedSchools() {
	f, err := os.ReadFile("./seeds/seed_schools.csv")
	if err != nil {
		log.Fatal(err)
		return
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

func testEndpointsSpeed() {
	filePaths := []string{
		// "features/admin/requests.http",      // Admin
		// "features/auth/requests.http",       // Auth
		// "features/hide_log/requests.http",   // Hide log
		"features/comments/requests.http",      // Comments
		"features/posts/requests.http",         // Posts
		"features/schools/requests.http",       // Schools
		"features/user/requests.http",          // Users
		"features/votes/requests.http",         // Votes
		"features/feedback/requests.http",      // Feedback
		"features/notifications/requests.http", // Notifications
		"features/saves/requests.http",         // Saves
		"features/reports/requests.http",       // Reports
		"features/drafts/requests.http",        // Drafts
		// Add more file paths here...
	}

	for _, filePath := range filePaths {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		// Convert content to string
		fileContent := string(content)

		// Define regular expressions for request parts

		re := regexp.MustCompile(`(?m)^(GET|POST|DELETE|PATCH|PUT) (http:\/\/[^\s]+)\n(.*(?:\n.+)*)\n\n\{([^}]*)\}$`)
		matches := re.FindAllStringSubmatch(fileContent, -1)

		for _, match := range matches {
			method := match[1]
			url := match[2]
			headers := match[3]
			body := match[4]

			// Parse headers
			headerRegex := regexp.MustCompile(`(?m)([^\n:]+):\s+(.*)`)
			headerMatches := headerRegex.FindAllStringSubmatch(headers, -1)

			headerMap := make(map[string]string)
			for _, headerMatch := range headerMatches {
				headerMap[headerMatch[1]] = headerMatch[2]
			}

			// Format for marshal
			body = "{" + body + "}"

			responseTime, err := makeRequest(method, url, headerMap, []byte(body))
			if err != nil {
				fmt.Printf("Error making request: %v\n", err)
				continue
			}

			if responseTime >= 200*time.Millisecond {
				fmt.Println("FAIL (>= 200ms)")
				fmt.Println("Method:", method)
				fmt.Println("URL:", url)
				fmt.Println("Body:", body)
				fmt.Printf("Response time: %s\n", responseTime)
				fmt.Println(strings.Repeat("-", 20))
			} else {
				fmt.Println("PASS (< 200ms)")
			}

		}
	}
}

func makeRequest(method, url string, headers map[string]string, body []byte) (time.Duration, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime)
	return responseTime, nil
}
