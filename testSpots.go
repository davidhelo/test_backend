package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
)

// struct that represents a spot
type Spot struct {
	Id        	string	`json:"id"`
	Name      	string 	`json:"name"`
	Website		sql.NullString  `json:"website"`
	Coordinates string	`json:"coordinates"`
	Description sql.NullString 	`json:"description"`
	Rating    	sql.NullFloat64 `json:"rating"`
	Distance  	float64 `json:"distance"`
}

func openDatabaseConnection() (DB *sql.DB, err error){
	//database connection parameters
	const (
		host     = "localhost"
		port     = 5432
	)
	 user     := os.Getenv("USERNAME_DB")
	 password := os.Getenv("USERPASSWORD_DB")
	 dbname   := os.Getenv("DB_NAME")

	psqlConnection := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Initialise database connection
	DB, err = sql.Open("postgres", psqlConnection)
	if err != nil {
		log.Fatal(err)
	}
	// defer DB.Close()

	errPing := DB.Ping()
	if errPing != nil {
		panic(errPing)
	}
	return DB, err
}

func main() {
	var err error
	//load .env file
	err = godotenv.Load(".env")
	if err != nil {
			log.Fatalf("Error loading environment variables file")
		}

	// Define the HTTP endpoint
	http.HandleFunc("/spots", findSpotsHandler)

	// Start the server
	http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func findSpotsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request parameters
	latitude, err := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	radius, err := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	if err != nil || radius <= 0 {
		http.Error(w, "Invalid radius", http.StatusBadRequest)
		return
	}

	shapeType := r.URL.Query().Get("shape")
	if shapeType != "circle" && shapeType != "square" {
		http.Error(w, "Invalid area type, valid values circle or square", http.StatusBadRequest)
		return
	}

	
	fmt.Println("REQUEST DATA: ", latitude, longitude, radius, shapeType)
	
	// Find spots in the specified area
	spots, err := findSpotsInArea(latitude, longitude, radius, shapeType)
	if err != nil {
		http.Error(w, "Failed to find spots", http.StatusInternalServerError)
		return
	}

	// Convert spots to JSON
	jsonData, err := json.Marshal(spots)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	// Set the response content type
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON data to the response
	w.Write(jsonData)
}


func findSpotsInArea(latitude, longitude, radius float64, shapeType string) ([]Spot, error) {
	var spots []Spot

	// convert radius meters to latitude/longitude degrees
	radiusInDegrees := radius / 111320.0 // considering 111320 meters to be distance per degree

	// Set shape subquery for shape
	shapeQuery := ""
	if shapeType == "circle" {
		shapeQuery += fmt.Sprintf(`ST_Buffer(ST_GeomFromText('POINT(%f %f)', 4326), %f, 'quad_segs=8')`, latitude, longitude, radiusInDegrees)
	} else { // else query for square shape
		shapeQuery += fmt.Sprintf(`ST_GeomFromText('POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))', 4326)`, 
			latitude - radiusInDegrees, longitude + radiusInDegrees, 
			latitude + radiusInDegrees, longitude + radiusInDegrees, 
			latitude + radiusInDegrees, longitude - radiusInDegrees, 
			latitude - radiusInDegrees, longitude - radiusInDegrees, 
			latitude - radiusInDegrees, longitude + radiusInDegrees)
	}

	query := fmt.Sprintf(`
			SELECT 
				spot_ID,
				name,
				website,
				coordinates,
				description,
				rating,
				distance_meters 
			FROM (
			-- subquery to get all item within shape, with center in (latitude longitude)
				SELECT 
					id AS spot_ID,
					name,
					website,
					coordinates,
					description,
					rating,
					ST_Within(
						ST_GeomFromText(ST_AsText(coordinates), 4326), 
						%s
					) AS pointIsInsideShape,
					ROUND( CAST(
						ST_DistanceSphere(
							ST_GeomFromText('POINT(%f %f)'), -- central point coordinate (lat lon)
							ST_GeomFromText(ST_AsText(coordinates))
						) AS numeric), 2) AS distance_meters
				FROM "SPOTS") AS foo
			WHERE pointIsInsideShape = true 
			ORDER BY 
			(distance_meters < 50) DESC,
			(CASE WHEN distance_meters < 50 THEN rating END) DESC,
			distance_meters ASC;`, shapeQuery, latitude, longitude)

			fmt.Println(query)
	
	//open database
	DB, err := openDatabaseConnection()

	//execute query
	rows, err := DB.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	//close database
	DB.Close()
	
	// populate the spots slice from rows data
	for rows.Next() {
		var spot Spot
		err := rows.Scan(&spot.Id, &spot.Name, &spot.Website, &spot.Coordinates, &spot.Description, &spot.Rating, &spot.Distance)
		if err != nil {
			panic(err)
		}
		spots = append(spots, spot)
	}

	return spots, nil
}
