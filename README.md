# test backend in Golang: Spots geolocalization

## files

### Task 1 **Query**: [taskQuery.sql](https://github.com/davidhelo/test_backend/blob/main/taskQuery.sql)

### Task 2 **Endpoint**: [testSpots.go](https://github.com/davidhelo/test_backend/blob/main/testSpots.go)

## Specifications

### Endpoint which returns spots in a circle or square area. 
Completed in Golang.

1. **Endpoint receives 4 parameters:**
    - Latitude
    - Longitude
    - Radius (in meters)
    - Type (circle or square)

2. **Find all spots in the table (spots.sql) within the received parameters.**

3. **Results by distance.**
    - If distance between two spots is smaller than 50m, then it is ordered by rating. 

4. Endpoint returns an array of objects (JSON) containing all fields in the data set.

### Query
- Return 3 columns: spot name, domain and count number for domain.
- It returns spots which have a domain with a count greater than 1.
- Change the website field, so it only contains the domain.
    - Example: https://domain.com/index.php → domain.com
- Returns the count of spots that have the same domain.

