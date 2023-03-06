library(rnoaa)
library(dplyr)
# define a function to get ISD data for a given year
get_isd_year <- function(station_id, year) {
  start_date <- paste0(year, "-01-01")
  end_date <- paste0(year, "-12-31")
  isd(station_id, start_date = start_date, end_date = end_date)
}
# get data for station with ID 725030-14732 from 2018 to 2020
years <- 2018:2020
dat_list <- lapply(years, get_isd_year, station_id = "725030-14732")
dat <- bind_rows(dat_list)
