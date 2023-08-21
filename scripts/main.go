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
	case "--test-endpoints-speed":
		testEndpointsSpeed()
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
				fmt.Println("Method:", method)
				fmt.Println("URL:", url)
				fmt.Println("Body:", body)
				fmt.Printf("Response time: %s\n", responseTime)
				fmt.Println(strings.Repeat("-", 20))
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
