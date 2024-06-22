package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	"github.com/samiam2013/basicauth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load ./.env: %v", err)
	}
	creds := os.Getenv("BASIC_AUTH_CREDS")
	if strings.TrimSpace(creds) == "" {
		log.Fatalf("Need BASIC_AUTH_CREDS env var, none set")
	}
	mapCreds := make(map[string]string)
	pairs := strings.Split(creds, ",")
	for _, pair := range pairs {
		strings := strings.Split(pair, ":")
		if len(strings) != 2 {
			log.Fatalf("Basic auth pair longer than 2 after split on ':'")
		}
		mapCreds[strings[0]] = strings[1]
	}
	basicAuth, err := basicauth.Builder(mapCreds, basicauth.WithUnsafeHTTP())
	if err != nil {
		log.Fatalf("Failed to build basicauth middleware: %v", err)
	}

	connectStr := "postgresql://grafana:grafana@localhost/grafana?sslmode=disable"
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /reportatmosphere", basicAuth(reportAtmosphereHandler(db)))
	// TODO add basicAuth to this
	mux.HandleFunc("GET /fishlighttimes", getFishLightTimes)

	_ = http.ListenAndServe(":8080", mux)
}

func reportAtmosphereHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ?temp=72.77&pressure=1006.29&humidity=54.25
		// get the temp, pressure, and humidity from the query params
		temp := r.URL.Query().Get("temp")
		// parse to float64
		tempF, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			http.Error(w, "Failed to parse temp as float64", http.StatusBadRequest)
			return
		}
		pressure := r.URL.Query().Get("pressure")
		pressureF, err := strconv.ParseFloat(pressure, 64)
		if err != nil {
			http.Error(w, "Failed to parse pressure as float64", http.StatusBadRequest)
			return
		}
		humidity := r.URL.Query().Get("humidity")
		humidityF, err := strconv.ParseFloat(humidity, 64)
		if err != nil {
			http.Error(w, "Failed to parse humidity as float64", http.StatusBadRequest)
			return
		}
		/*grafana=# select * from sensor_data limit 0;
		 id | created_at | temperature | pressure | humidity
		----+------------+-------------+----------+----------
		(0 rows)*/
		_, err = db.Exec(
			"INSERT INTO sensor_data (temperature, pressure, humidity) VALUES ($1, $2, $3)",
			tempF, pressureF, humidityF)
		if err != nil {
			http.Error(w, "Failed to insert sensor data", http.StatusInternalServerError)
			log.Printf("Failed to insert sensor data: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK; sensor data inserted"))
	}
}

func getFishLightTimes(w http.ResponseWriter, r *http.Request) {
	path := os.Getenv("FISH_LIGHT_TIMES_FILE_PATH")
	if strings.TrimSpace(path) == "" {
		http.Error(w, "FISH_LIGHT_TIMES_FILE_PATH env var required, not set", http.StatusInternalServerError)
		return
	}
	fh, err := os.Open(path)
	if err != nil {
		http.Error(w, "Failed to open fight light times file", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, fh); err != nil {
		http.Error(w, "Failed to write fish light times from file", http.StatusInternalServerError)
		return
	}
}
