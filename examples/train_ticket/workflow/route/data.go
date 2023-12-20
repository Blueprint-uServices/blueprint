package route

type Route struct {
	ID           string
	Stations     []string
	Distances    []int64
	StartStation string
	EndStation   string
}

type RouteInfo struct {
	ID           string
	StartStation string
	EndStation   string
	StationList  string
	DistanceList string
}
