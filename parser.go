package addresser

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"github.com/damonchen/collection"
)

//go:embed data/states.json
var _allStates []byte

//go:embed data/us-street-types.json
var _usStreetTypes []byte

//go:embed data/cities.json
var _allCities []byte

//go:embed data/us-states.json
var _usStates []byte

//go:embed data/us-cities.json
var _usCities []byte

func getAllStates() (map[string]string, error) {
	var r map[string]string
	err := json.Unmarshal(_allStates, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getUsStreetTypes() (map[string]string, error) {
	var r map[string]string
	err := json.Unmarshal(_usStreetTypes, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getAllCities() (map[string][]string, error) {
	var r map[string][]string
	err := json.Unmarshal(_allCities, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getUsStates() (map[string]string, error) {
	var r map[string]string
	err := json.Unmarshal(_usStates, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getUsCities() (map[string][]string, error) {
	var r map[string][]string
	err := json.Unmarshal(_usCities, &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func toTitleCase(str string) string {
	pattern := regexp.MustCompile(`\w\S*`)
	return pattern.ReplaceAllStringFunc(str, func(txt string) string {
		return strings.ToUpper(txt[0:1]) + strings.ToLower(txt[1:])
	})
}

func getKeyByValue[T comparable](object map[string]T, value T) string {
	for key, v := range object {
		if v == value {
			return key
		}
	}
	return ""
}

var usStreetDirectional = map[string]string{
	"north":     "N",
	"northeast": "NE",
	"east":      "E",
	"southeast": "SE",
	"south":     "S",
	"southwest": "SW",
	"west":      "W",
	"northwest": "NW",
}

var usLine2Prefixes = map[string]string{
	"APARTMENT":  "APT",
	"APT":        "APT",
	"BASEMENT":   "BSMT",
	"BSMT":       "BSMT",
	"BLDG":       "BLDG",
	"BUILDING":   "BLDG",
	"DEPARTMENT": "DEPT",
	"DEPT":       "DEPT",
	"FL":         "FL",
	"FLOOR":      "FL",
	"FRNT":       "FRNT",
	"FRONT":      "FRNT",
	"HANGAR":     "HNGR",
	"HNGR":       "HNGR",
	"LBBY":       "LBBY",
	"LOBBY":      "LBBY",
	"LOT":        "LOT",
	"LOWER":      "LOWR",
	"LOWR":       "LOWER",
	"OFC":        "OFC",
	"OFFICE":     "OFC",
	"PENTHOUSE":  "PH",
	"PH":         "PH",
	"PIER":       "PIER",
	"REAR":       "REAR",
	"RM":         "RM",
	"ROOM":       "RM",
	"SIDE":       "SIDE",
	"SLIP":       "SLIP",
	"SPACE":      "SPC",
	"SPC":        "SPC",
	"STE":        "STE",
	"STOP":       "STOP",
	"SUITE":      "STE",
	"TRAILER":    "TRLR",
	"TRLR":       "TRLR",
	"UNIT":       "UNIT",
	"UPPER":      "UPPR",
	"UPPR":       "UPPR",
	"#":          "#",
}

func replaceRepeatSpace(s string) string {
	pattern := regexp.MustCompile(" +")
	return pattern.ReplaceAllString(s, "")
}

func splitAddress(s string, p string) []string {
	pattern := regexp.MustCompile(p)
	return pattern.Split(s, -1)
}

type Address struct {
	ID                string `json:"id"`
	ZipCode           string `json:"zipCode"`
	ZipCodePlusFour   string `json:"zipCodePlusFour"`
	StateAbbreviation string `json:"stateAbbreviation"`
	StateName         string `json:"stateName"`
	PlaceName         string `json:"placeName"`
	AddressLine1      string `json:"addressLine1"`
	AddressLine2      string `json:"addressLine2"`
	StreetNumber      string `json:"streetNumber"`
	FormattedAddress  string `json:"formattedAddress"`
	StreetDirection   string `json:"streetDirection"`
	StreetName        string `json:"streetName"`
	StreetSuffix      string `json:"streetSuffix"`
}

func (a Address) GoString() string {
	data, _ := json.MarshalIndent(a, "", "  ")
	return fmt.Sprintf("%s", string(data))
}

func ParseAddress(address string) (*Address, error) {
	if len(address) == 0 {
		return nil, errors.New("argument must be a non-empty string")
	}

	// Deal with any repeated spaces
	pattern := regexp.MustCompile(` +`)
	address = pattern.ReplaceAllString(address, " ")

	// Assume comma, newline and tab is an intentional delimiter
	pattern = regexp.MustCompile(`,|\\t|\\n`)
	addressParts := pattern.Split(address, -1)

	r := Address{}

	// Check if the last section contains country reference (Just supports US for now)
	countrySection := strings.TrimSpace(addressParts[len(addressParts)-1])
	if countrySection == "US" || countrySection == "USA" || countrySection == "United States" || countrySection == "Canada" {
		addressParts = addressParts[:len(addressParts)-1]
	}

	// Assume the last address section contains state, zip or both
	stateString := strings.TrimSpace(addressParts[len(addressParts)-1])

	// Parse and remove zip or zip plus 4 from end of string
	pattern = regexp.MustCompile(`\d{5}$`)
	pattern2 := regexp.MustCompile(`\d{5}-\d{4}$`)
	pattern3 := regexp.MustCompile(`[A-Z]\d[A-Z] ?\d[A-Z]\d`)
	if pattern.MatchString(stateString) {
		matches := pattern.FindStringSubmatch(stateString)
		r.ZipCode = matches[0]
		stateString = strings.TrimSpace(stateString[0 : len(stateString)-5])
	} else if pattern2.MatchString(stateString) {
		matches := pattern2.FindStringSubmatch(stateString)
		zipString := matches[0]
		zipCode := zipString[:5]

		r.ZipCode = zipCode
		r.ZipCodePlusFour = zipString

		stateString = strings.TrimSpace(stateString[0 : len(stateString)-10])
	} else if pattern3.MatchString(stateString) {
		matches := pattern3.FindStringSubmatch(stateString)
		zipCode := matches[0]
		r.ZipCode = zipCode
		stateString = strings.TrimSpace(stateString[0 : len(stateString)-len(zipCode)])
	}

	// Parse and remove state
	if len(stateString) > 0 {
		addressParts[len(addressParts)-1] = stateString
	} else {
		addressParts = addressParts[0 : len(addressParts)-1]
		stateString = strings.TrimSpace(addressParts[len(addressParts)-1])
	}

	// First check for just an Abbreviation
	allStates, _ := getAllStates()

	key := getKeyByValue(allStates, strings.ToUpper(stateString))

	if len(stateString) == 2 && key != "" {
		r.StateAbbreviation = strings.ToUpper(stateString)
		r.StateName = toTitleCase(key)
		stateString = stateString[0 : len(stateString)-2]
	} else {
		for key := range allStates {
			pattern := regexp.MustCompile(fmt.Sprintf("(?i) %s$|%s$", allStates[key], key))
			if pattern.MatchString(stateString) {
				stateString = pattern.ReplaceAllString(stateString, "")
				r.StateAbbreviation = allStates[key]
				r.StateName = toTitleCase(key)
				break
			}
		}
	}

	if len(r.StateAbbreviation) == 0 || len(r.StateAbbreviation) != 2 {
		return nil, errors.New("can not parse address. State not found")
	}

	// Parse and remove city/place name
	placeString := ""
	if len(stateString) > 0 { // Check if anything is left of last section
		addressParts[len(addressParts)-1] = stateString
		placeString = addressParts[len(addressParts)-1]
	} else {
		addressParts = addressParts[:len(addressParts)-1]
		placeString = strings.TrimSpace(addressParts[len(addressParts)-1])
	}

	allCities, _ := getAllCities()
	r.PlaceName = ""
	for _, element := range allCities[r.StateAbbreviation] {
		re := regexp.MustCompile("(?i)" + element + "$") /// 需要忽略大小写
		if re.MatchString(placeString) {
			placeString = re.ReplaceAllString(placeString, "") // Carve off the place name

			r.PlaceName = element
			break // Found a winner - stop looking for cities
		}
	}

	if len(r.PlaceName) != 0 {
		r.PlaceName = toTitleCase(r.PlaceName)
		placeString = ""
	}

	// Parse the street data
	streetString := ""
	usStreetDirectionalString := strings.Join(collection.Values(usStreetDirectional), "|")
	usLine2String := strings.Join(collection.Keys(usLine2Prefixes), "|")

	if len(placeString) > 0 { // Check if anything is left of last section
		addressParts[len(addressParts)-1] = placeString
	} else {
		addressParts = addressParts[0 : len(addressParts)-1]
	}

	if len(addressParts) > 2 {
		return nil, errors.New("can not parse address. More than two address lines")
	} else if len(addressParts) == 2 {
		// check if the secondary data is first
		pattern := regexp.MustCompile(fmt.Sprintf("(?i)^(%s)\\\\b", usLine2String))
		if pattern.MatchString(addressParts[0]) {
			tmpString := addressParts[1]
			addressParts[1] = addressParts[0]
			addressParts[0] = tmpString
		}

		//Assume street line is first
		r.AddressLine2 = strings.TrimSpace(addressParts[1])
		addressParts = addressParts[0 : len(addressParts)-1]
	}

	if len(addressParts) == 1 {
		streetString = strings.TrimSpace(addressParts[0])
		// If no address line 2 exists check to see if it is incorrectly placed at the front of line 1
		if len(r.AddressLine2) == 0 {
			pattern := regexp.MustCompile(fmt.Sprintf("(?i)^(%s)\\\\s\\\\S+", usLine2String))
			if pattern.MatchString(streetString) {
				matches := pattern.FindStringSubmatch(stateString)
				r.AddressLine2 = matches[0]
				streetString = strings.TrimSpace(pattern.ReplaceAllString(streetString, ""))
			}
		}

		usStreetTypes, _ := getUsStreetTypes()

		//Assume street address comes first and the rest is secondary address
		reStreet := regexp.MustCompile(fmt.Sprintf("(?i)\\.\\*\\\\b(?:%s)\\\\b\\\\.?( +(?:%s)\\\\b)?",
			strings.Join(collection.Keys(usStreetTypes), "|"), usStreetDirectionalString))

		rePO := regexp.MustCompile(`(?i)(P\\.?O\\.?|POST\\s+OFFICE)\\s+(BOX|DRAWER)\\s\\w+`)
		reAveLetter := regexp.MustCompile(`(?i)\.\*\\b(ave.?|avenue)\.\*\\b[a-zA-Z]\\b`)
		reNoSuffix := regexp.MustCompile(`(?i)\b\d+\s[a-zA-Z0-9_ ]+\b`)

		if reAveLetter.MatchString(streetString) {
			r.AddressLine1 = reAveLetter.FindStringSubmatch(streetString)[0]
			streetString = strings.TrimSpace(reAveLetter.ReplaceAllString(streetString, "")) // Carve off the first address line

			if len(streetString) > 0 {
				if len(r.AddressLine2) > 0 {
					return nil, fmt.Errorf("can not parse address. Too many address lines. Input string: %s", address)
				} else {
					r.AddressLine2 = streetString
				}
			}

			streetParts := strings.Split(r.AddressLine1, " ")
			// Assume type is last and number is first
			r.StreetNumber = streetParts[0] // Assume number is first element

			// Normalize to Ave
			pattern := regexp.MustCompile("(?i)^(ave.?|avenue)$")
			streetParts[len(streetParts)-2] = pattern.ReplaceAllString(streetParts[len(streetParts)-2], "Ave")

			r.StreetName = streetParts[1] // Assume street name is everything in the middle
			for i := 2; i <= len(streetParts)-1; i++ {
				r.StreetName = r.StreetName + " " + streetParts[i]
			}

			r.StreetName = toTitleCase(r.StreetName)
			r.AddressLine1 = strings.Join([]string{r.StreetNumber, r.StreetName}, " ")
		} else if reStreet.MatchString(streetString) {
			r.AddressLine1 = pattern.FindStringSubmatch(streetString)[0]
			streetString = pattern.ReplaceAllString(streetString, "")
			if len(streetString) > 0 {
				if len(r.AddressLine2) > 0 {
					return nil, errors.New("can not parse address. Too many address lines. Input string: " + address)
				} else {
					r.AddressLine2 = streetString
				}
			}
			streetParts := strings.Split(r.AddressLine1, " ")

			// Check if directional is last element
			pattern = regexp.MustCompile(fmt.Sprintf("(?i)\\.\\*\\\\b(?:%s)$", usStreetDirectionalString))
			if pattern.MatchString(r.AddressLine1) {
				r.StreetDirection = strings.ToUpper(streetParts[len(streetParts)-1])
				streetParts = streetParts[0 : len(streetParts)-1]
			}

			usStreetTypes, _ := getUsStreetTypes()

			// Assume type is last and number is first
			r.StreetNumber = streetParts[0] // Assume number is first element

			// If there are only 2 street parts (number and name) then its likely missing a "real" suffix and the street name just happened to match a suffix
			if len(streetParts) > 2 {
				// Remove '.' if it follows streetSuffix
				pattern = regexp.MustCompile(`\.$`)
				streetParts[len(streetParts)-1] = pattern.ReplaceAllString(streetParts[len(streetParts)-1], "")
				r.StreetSuffix = toTitleCase(strings.ToLower(usStreetTypes[streetParts[len(streetParts)-1]]))
			}

			r.StreetName = streetParts[1] // Assume street name is everything in the middle
			for i := 2; i < len(streetParts)-1; i++ {
				r.StreetName = r.StreetName + " " + streetParts[i]
			}
			r.StreetName = toTitleCase(r.StreetName)
			r.AddressLine1 = strings.Join([]string{r.StreetNumber, r.StreetName}, " ")

			if len(r.StreetSuffix) > 0 {
				r.AddressLine1 = r.AddressLine1 + " " + r.StreetSuffix
			}

			if len(r.StreetDirection) > 0 {
				r.AddressLine1 = r.AddressLine1 + " " + r.StreetDirection
			}
		} else if rePO.MatchString(streetString) {
			r.AddressLine1 = rePO.FindStringSubmatch(streetString)[0]
			streetString = strings.TrimSpace(rePO.ReplaceAllString(streetString, ""))
		} else if reNoSuffix.MatchString(streetString) {
			// Check for a line2 prefix followed by a single word. If found peel that off as addressLine2
			reLine2 := regexp.MustCompile(fmt.Sprintf("(?i)\\\\s(%s)\\\\.?\\\\s[a-zA-Z0-9_\\-]+$", usLine2String))

			if reLine2.MatchString(streetString) {
				r.AddressLine2 = strings.TrimSpace(reLine2.FindStringSubmatch(streetString)[0])
				streetString = strings.TrimSpace(reLine2.ReplaceAllString(streetString, "")) // Carve off the first address line
			}

			r.AddressLine1 = reNoSuffix.FindStringSubmatch(streetString)[0]
			streetString = strings.TrimSpace(reNoSuffix.ReplaceAllString(streetString, "")) // Carve off the first address line

			streetParts := strings.Split(r.AddressLine1, " ")

			// Assume type is last and number is first
			r.StreetNumber = streetParts[0]               // Assume number is first element
			streetParts = streetParts[1:]                 // Remove the first element
			r.StreetName = strings.Join(streetParts, " ") // Assume street name is everything else
		} else {
			return nil, errors.New("can not parse address. Invalid street address data. Input string: " + address)
		}
	} else {
		return nil, errors.New("can not parse address. Invalid street address data. Input string: " + address)
	}

	addressString := r.AddressLine1
	if len(r.AddressLine2) > 0 {
		addressString += ", " + r.AddressLine2
	}

	if len(addressString) > 0 && len(r.PlaceName) > 0 && len(r.StateAbbreviation) > 0 && len(r.ZipCode) > 0 {
		idString := addressString + ", " + r.PlaceName + ", " + r.StateAbbreviation + " " + r.ZipCode
		r.FormattedAddress = idString
		ID := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
			strings.ReplaceAll(idString, " ", "-"),
			"#", "-"), "//", "-"), ".", "-",
		)
		r.ID = url.QueryEscape(ID)
	}

	return &r, nil

}

func randomProperty[T any](obj map[string]T) string {
	var keys []string
	for key := range obj {
		keys = append(keys, key)
	}

	return keys[len(keys)*rand.Int()<<0]
}

type City struct {
	City  string
	State string
}

func RandomCity() (City, error) {
	usCities, err := getUsCities()
	if err != nil {
		return City{}, err
	}
	randomState := randomProperty(usCities)
	randomStateData := usCities[randomState]
	randomCityElementId := rand.Intn(len(randomStateData))
	randomCity := randomStateData[randomCityElementId]

	return City{
		City:  randomCity,
		State: randomState,
	}, nil
}

func Cities() (map[string][]string, error) {
	return getUsCities()
}
