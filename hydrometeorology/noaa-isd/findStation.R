# Load required libraries
library(shiny)
library(dplyr)
library(geosphere)

# Load station list
if(!file.exists("isd-history.csv")) {
  download.file("https://www.ncei.noaa.gov/pub/data/noaa/isd-history.csv", "isd-history.csv")
}
stations <- read.csv("isd-history.csv",
                     header = TRUE, stringsAsFactors = FALSE)

# Define the ui
ui <- fluidPage(
  titlePanel("Find NOAA-ISD stations within a certain range"),
  
  sidebarLayout(
    sidebarPanel(
      numericInput("lat", "Enter latitude:", value = 0, step = 0.01),
      numericInput("lon", "Enter longitude:", value = 0, step = 0.01),
      numericInput("range", "Enter range (km):", value = 0, step = 0.01),
      actionButton("submit", "Submit")
    ),
    
    mainPanel(
      h3("Stations within range:"),
      tableOutput("station_list")
    )
  )
)

# Define the server
server <- function(input, output, session) {
  
  # Define a reactive expression to find the stations within range
  stations_within_range <- reactive({
    
    # Calculate the distances between the input coordinates and the coordinates of each station
    distances <- distHaversine(stations[, c("LON", "LAT")], c(input$lon, input$lat))
    distances_km <- distances/1000
    
    # Subset the stations data to include only the stations within the input range
    stations_within_range <- stations[distances_km <= input$range, c("USAF", "WBAN", "STATION.NAME", "ICAO", "LAT", "LON", "ELEV.M.", "BEGIN", "END")]
    stations_within_range$AVAILABLE.TIME.RANGE <- paste(stations_within_range$BEGIN, stations_within_range$END, sep = " - ")
    stations_within_range$BEGIN <- NULL
    stations_within_range$END <- NULL
    
    # Add the distances to the output data frame
    stations_within_range$distance_km <- round(distances_km[distances_km <= input$range], 2)
    
    # Filter out NA records
    stations_within_range <- stations_within_range[!is.na(stations_within_range$distance_km), ]
    
    # Sort the data frame by distance_km
    stations_within_range <- arrange(stations_within_range, distance_km)
    
    stations_within_range
  })
  
  # Define the output
  output$station_list <- renderTable({
    req(input$submit)
    stations_within_range()
  })
}

# Run the app
shinyApp(ui, server)
