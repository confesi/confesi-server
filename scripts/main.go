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
	"path/filepath"
	"regexp"
	"runtime"
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
	case "--seed-award-types":
		seedAwardTypes()
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
	seedAwardTypes()
}

func seedAwardTypes() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_award_types.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	// Initialize a slice to store the award types
	awardTypes := []db.AwardType{}
	reader.Read() // Skip the header

	// Read and parse each line of the CSV file
	for {
		result, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading CSV: %v", err)
		}

		// Create an AwardType object and append it to the slice
		awardType := db.AwardType{
			Name:        result[0],
			Description: result[1],
			Icon:        result[2],
		}
		awardTypes = append(awardTypes, awardType)
	}

	// Connect to the database and insert the award types
	pg := db.New()
	result := pg.Model(&[]db.AwardType{}).Clauses(clause.OnConflict{DoNothing: true}).Create(awardTypes)
	if result.Error != nil {
		log.Fatalf("Failed to seed award types: %v", result.Error)
	}

	fmt.Println("Seeding Award Types Completed")
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(awardTypes))
}

func seedYearOfStudies() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_year_of_study.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	yearOfStudies := []db.YearOfStudy{}
	reader.Read() // Skip the header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
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
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(yearOfStudies))
}

func seedFaculties() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_faculties.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	faculties := []db.Faculty{}
	reader.Read() // Skip the header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
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
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(faculties))
}

func seedPostCategories() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_post_categories.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	categories := []db.PostCategory{}
	reader.Read() // Skip the header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
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
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(categories))
}

func seedReportTypes() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_report_types.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	types := []db.ReportType{}
	reader.Read() // Skip the header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}

		reportType := db.ReportType{
			Type: result[0],
		}

		types = append(types, reportType)
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
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(types))
}

func seedFeedbackTypes() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_feedback_types.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
		return
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	types := []db.FeedbackType{}
	reader.Read() // Skip the header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}

		feedbackType := db.FeedbackType{
			Type: result[0],
		}

		types = append(types, feedbackType)
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
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(types))
}

func seedSchools() {
	// Determine the base path of the current script
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)

	// Construct the full path to the seed file
	seedFilePath := filepath.Join(basePath, "../seeds/seed_schools.csv")

	// Read the file
	f, err := os.ReadFile(seedFilePath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}
	buf := bytes.NewReader(f)
	reader := csv.NewReader(buf)

	schools := []db.School{}
	reader.Read() // Skip the header

	for {
		result, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
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

	result := pg.
		Model(&[]db.School{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(schools)

	if result.Error != nil {
		panic(result.Error.Error())
	}
	fmt.Println("Seeding Schools Completed")
	fmt.Printf("Rows Updated: %v / %v\n", result.RowsAffected, len(schools))
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
		// "handlers/admin/requests.http",      // Admin
		// "handlers/auth/requests.http",       // Auth
		// "handlers/hide_log/requests.http",   // Hide log
		"handlers/comments/requests.http",      // Comments
		"handlers/posts/requests.http",         // Posts
		"handlers/schools/requests.http",       // Schools
		"handlers/user/requests.http",          // Users
		"handlers/votes/requests.http",         // Votes
		"handlers/feedback/requests.http",      // Feedback
		"handlers/notifications/requests.http", // Notifications
		"handlers/saves/requests.http",         // Saves
		"handlers/reports/requests.http",       // Reports
		"handlers/drafts/requests.http",        // Drafts
		// todo: add more file paths here...
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
