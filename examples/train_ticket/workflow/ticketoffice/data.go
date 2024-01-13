package ticketoffice

type Office struct {
	OfficeName string
	Address    string
	WorkTime   string
	WindowNum  uint16
}

type InnerRegion struct {
	Name string `json:"region"`
}

type City struct {
	Name         string        `json:"city"`
	InnerRegions []InnerRegion `json:"regions"`
}

type Region struct {
	Province string `json:"province"`
	Cities   []City `json:"cities"`
}
