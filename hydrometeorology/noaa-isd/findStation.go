package main

import (
    "fmt"
    "github.com/tidwall/gjson"
    "io/ioutil"
    "math"
    "net/http"
    "strconv"
    "strings"

    "github.com/golang/protobuf/proto"
    "github.com/gorilla/mux"
    "github.com/shiny"
)

type Station struct {
    USAF            string
    WBAN            string
    StationName     string
    ICAO            string
    Latitude        float64
    Longitude       float64
    Elevation       float64
    Begin           string
    End             string
    AvailableTime   string
    Distance_km     float64
}

func main() {
    // Load station list
    if _, err := os.Stat("isd-history.csv"); os.IsNotExist(err) {
        downloadFile("https://www.ncei.noaa.gov/pub/data/noaa/isd-history.csv", "isd-history.csv")
    }
    stations, _ := loadStations("isd-history.csv")

    // Define the router
    router := mux.NewRouter()
    router.HandleFunc("/", index)
    router.HandleFunc("/stations", getStationsInRange(stations)).Methods("POST")

    // Run the app
    http.ListenAndServe(":8080", router)
}

func index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to the NOAA-ISD station finder!")
}

func getStationsInRange(stations []Station) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get the input parameters
        lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
        lon, _ := strconv.ParseFloat(r.FormValue("lon"), 64)
        range_km, _ := strconv.ParseFloat(r.FormValue("range"), 64)

        // Find the stations within range
        stations_within_range := make([]Station, 0)
        for _, station := range stations {
            // Calculate the distance between the input coordinates and the coordinates of each station
            distance_km := distance(lat, lon, station.Latitude, station.Longitude)

            // Add the station to the output array if it is within the input range
            if distance_km <= range_km {
                station.Distance_km = distance_km
                station.AvailableTime = fmt.Sprintf("%s - %s", station.Begin, station.End)
                stations_within_range = append(stations_within_range, station)
            }
        }

        // Sort the stations by distance
        sortByDistance(stations_within_range)

        // Write the output
        b, err := proto.Marshal(&StationList{Stations: stations_within_range})
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/octet-stream")
        w.Write(b)
    }
}

func downloadFile(url string, filename string) {
    resp, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile(filename, body, 0644)
    if err != nil {
        panic(err)
    }
}

func loadStations(filename string) ([]Station, error) {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // Parse the CSV data into an array of Station structs
    data := string(bytes)
    lines := strings.Split(data, "\n")
    stations := make([]Station, len(lines)-1)
    for i, line := range lines[1:] {
        fields := strings.Split(line, ",")
        stations[i] = Station{
            USAF:        fields[0],
            WBAN:        fields[1],
            StationName: fields[2],
            ICAO:        fields[3],
            Latitude:    parseCoord(fields[6]),
            Longitude:   parseCoord(fields[7]),
            Elevation:   parseElevation(fields[8]),
            Begin:       fields[10],
            End:         fields[11],
        }
    }
    return stations, nil
}

func parseCoord(coord string) float64 {
    degrees, _ := strconv.ParseFloat(coord[0:3], 64)
    minutes, _ := strconv.ParseFloat(coord[3:5], 64)
    seconds, _ := strconv.ParseFloat(coord[5:7], 64)
    direction := coord[7:8]
    decimal := degrees + (minutes / 60) + (seconds / 3600)
    if direction == "S" || direction == "W" {
        decimal = -decimal
    }
    return decimal
}

func parseElevation(elevation string) float64 {
    if elevation == "" {
        return 0
    }
    value, _ := strconv.ParseFloat(elevation, 64)
    return value
}

func distance(lat1, lon1, lat2, lon2 float64) float64 {
    // Calculate the distance between two points on the Earth's surface using the Haversine formula
    r := 6371.0 // Earth's radius in km
    dLat := deg2rad(lat2 - lat1)
    dLon := deg2rad(lon2 - lon1)
    a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return r * c
}

func deg2rad(deg float64) float64 {
    return deg * (math.Pi / 180)
}

func sortByDistance(stations []Station) {
    // Sort the stations by distance
    less := func(i, j int) bool {
        return stations[i].Distance_km < stations[j].Distance_km
    }
    sort.Slice(stations, less)
}
