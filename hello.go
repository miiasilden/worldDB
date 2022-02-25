package main

import (
    "fmt"
    "log"
    "encoding/json"
    "net/http"
    "sync"
    "strconv"
    "sort"
    "strings"
    "io/ioutil"
    "errors"
    "github.com/gorilla/mux"
)

//JSON and search map structs
type ContCountryCity struct {
    Continent string
    Country string
    City string
}

type ContCountry struct {
    Continent string
    Country string
}

type Continent struct {
    Continent string
}

type Country struct {
    Country string
}

type City struct {
    City string
}

type ChangeName struct {
    OldName string
    NewName string
}

//DB record structs
type ContinentRecord struct {
    Id int
    Name string
    NamePtr *string
}

type CountryRecord struct {
    Id int
    Name string
    NamePtr *string
    Continent *string
}

type CityRecord struct {
    Id int
    Name string
    NamePtr *string
    Country *string
    Continent *string
}

var continentId int
var freeContIdList []int
var countryId int
var freeCountryIdList []int
var cityId int
var freeCityIdList []int
var mutex = &sync.Mutex{}
// ID based search trees
var keyContCountryValCities map[string][]*string
var keyContValCities map[int][]*string
var keyCountryValCities map[int][]*string
var keyContValCountries map[int][]*string
//DB - name based search trees (to find ID)
var keyContinentValId map[string]ContinentRecord
var keyCountryValId map[string]CountryRecord
var keyCityValId map[string]CityRecord

func homePage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to worldDB\n")
    fmt.Fprintf(w, "With this tool you can create, delete, modify and query continents, countries and cities\n")
    fmt.Fprintf(w, "For help: localhost:10000/help\n")
}

func helpPage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "localhost:10000/help\n")
    fmt.Fprintf(w, "      get this help\n")
    fmt.Fprintf(w, "GET localhost:10000/continents\n")
    fmt.Fprintf(w, "      reads and lists all continents in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/countries\n")
    fmt.Fprintf(w, "      reads and lists all countries in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/cities\n")
    fmt.Fprintf(w, "      reads and lists all cities in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/country/info + json data: \"Country\":\"<country name>\"\n")
    fmt.Fprintf(w, "      reads and lists all countries in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/city/info + json data: \"City\":\"<city name>\"\n")
    fmt.Fprintf(w, "      reads and lists all cities in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/continent/countries + json data: \"Continent\":\"<cont name>\"\n")
    fmt.Fprintf(w, "      reads and lists all countries on continent <cont name> in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/continent/country/cities + json data: \"Continent\":\"<cont name>\",\"Country\":\"<country name>\"\n")
    fmt.Fprintf(w, "      reads and lists all cities of country <country name> on continent <cont name> in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/continent/cities + json data: \"Continent\":\"<cont name>\"\n")
    fmt.Fprintf(w, "      reads and lists all cities on continent <cont name> in DB\n")
    fmt.Fprintf(w, "GET localhost:10000/country/cities + json data: \"Country\":\"<country name>\"\n")
    fmt.Fprintf(w, "      reads and lists all cities of country <country name> in DB\n")
    fmt.Fprintf(w, "POST localhost:10000/continent + json data: \"Continent\":\"<cont name>\"\n")
    fmt.Fprintf(w, "      create a continent into DB\n")
    fmt.Fprintf(w, "POST localhost:10000/country + json data: \"Continent\":\"<cont name>\",\"Country\":\"<country name>\"\n")
    fmt.Fprintf(w, "      create a country into DB\n")
    fmt.Fprintf(w, "POST localhost:10000/city + json data: \"Continent\":\"<cont name>\",\"Country\":\"<country name>\",\"City\":\"<city name>\"\n")
    fmt.Fprintf(w, "      create a city into DB\n")
    fmt.Fprintf(w, "DELETE localhost:10000/continent + json data: \"Continent\":\"<cont name>\"\n")
    fmt.Fprintf(w, "      delete a continent from DB.\n")
    fmt.Fprintf(w, "      Note! Delete of continent will delete also all the countries and cities of that continent from whole DB\n")
    fmt.Fprintf(w, "DELETE localhost:10000/country + json data: \"Country\":\"<country name>\"\n")
    fmt.Fprintf(w, "      delete a country from DB.\n")
    fmt.Fprintf(w, "      Note! Delete of country will delete also all the cities of that country from whole DB\n")
    fmt.Fprintf(w, "DELETE localhost:10000/city + json data: \"City\":\"<city name>\"\n")
    fmt.Fprintf(w, "      delete a city from DB\n")
    fmt.Fprintf(w, "PUT localhost:10000/continent/name + json data: \"Old name\":\"<old cont name>\"\"New name\":\"<new cont name>\"\n")
    fmt.Fprintf(w, "      update a continent name in DB\n")
    fmt.Fprintf(w, "PUT localhost:10000/country/name + json data: \"Old name\":\"<old country name>\"\"New name\":\"<new country name>\"\n")
    fmt.Fprintf(w, "      update a country name in DB\n")
    fmt.Fprintf(w, "PUT localhost:10000/country + json data: \"Continent\":\"<cont name>\",\"Country\":\"<country name>\"\n")
    fmt.Fprintf(w, "      update a country in DB\n")
    fmt.Fprintf(w, "PUT localhost:10000/city/name + json data: \"Old name\":\"<old city name>\"\"New name\":\"<new city name>\"\n")
    fmt.Fprintf(w, "      update a city name in DB\n")
    fmt.Fprintf(w, "PUT localhost:10000/city + json data: \"Continent\":\"<cont name>\",\"Country\":\"<country name>\",\"City\":\"<city name>\"\n")
    fmt.Fprintf(w, "      update a city in DB\n")

}

func addToFreeList(freeList *[]int, id int) {
    *freeList = append(*freeList, id)
}

func getNewId (idType string) int {
    var newId int
    switch idType {
        case "city":
            if len(freeCityIdList) > 0 {
		newId = freeCityIdList[len(freeCityIdList)-1]
		mutex.Lock()
		freeCityIdList = freeCityIdList[:len(freeCityIdList)-1]
		mutex.Unlock()
		fmt.Println(freeCityIdList)
	    } else {
		incrId(&cityId)
		newId = cityId
	    }
        case "country":
            if len(freeCountryIdList) > 0 {
		newId = freeCountryIdList[len(freeCountryIdList)-1]
		mutex.Lock()
		freeCountryIdList = freeCountryIdList[:len(freeCountryIdList)-1]
		mutex.Unlock()
		fmt.Println(freeCountryIdList)
	    } else {
		incrId(&countryId)
		newId = countryId
	    }
        case "continent":
            if len(freeContIdList) > 0 {
		newId = freeContIdList[len(freeContIdList)-1]
		mutex.Lock()
		freeContIdList = freeContIdList[:len(freeContIdList)-1]
		mutex.Unlock()
		fmt.Println(freeContIdList)
	    } else {
		incrId(&continentId)
		newId = continentId
	    }
    }
    return newId
}

func incrId(id *int) {
    mutex.Lock()
    *id++
    mutex.Unlock()
}

func deleteCityDbEntry(city string) {
    cityRecord := keyCityValId[city]
    addToFreeList(&freeCityIdList, cityRecord.Id)
    //delete DB entry
    delete(keyCityValId, city)
}

func deleteCityDbEntryWithId(city string, id int) {
    addToFreeList(&freeCityIdList, id)
    //delete DB entry
    delete(keyCityValId, city)
}

func deleteCountryDbEntryWithId(country string, id int) {
    addToFreeList(&freeCountryIdList, id)
    //delete DB entry
    delete(keyCountryValId, country)
}

func deleteContDbEntryWithId(cont string, id int) {
    addToFreeList(&freeContIdList, id)
    //delete DB entry
    delete(keyContinentValId, cont)
}

func parseContCountryCityJson(w http.ResponseWriter, r *http.Request, city *ContCountryCity) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
	return err
    }
    err = json.Unmarshal(body, city)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
	return err
    }
    if city.City == "" || city.Country == ""|| city.Continent == "" {
        fmt.Fprintf(w, "not enough info - json Continent:" + city.Continent + ", Country:" + city.Country + ", City:" + city.City + "\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
	return errors.New("Not enough info")
    }
    return err
}

func parseContCountryJson(w http.ResponseWriter, r *http.Request, country *ContCountry) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
    }
    err = json.Unmarshal(body, country)
    if err != nil {
        fmt.Fprint(w, err.Error())
       panic(err)
    }
    if (country.Continent == "" || country.Country == "") {
        fmt.Fprintf(w, "not enough info - json Continent:" + country.Continent + ", Country:" + country.Country + "\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
	return errors.New("Not enough info")
    }
    return err
}
func parseContJson(w http.ResponseWriter, r *http.Request, cont *Continent) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
    }
    err = json.Unmarshal(body, cont)
    if err != nil {
        fmt.Fprint(w, err.Error())
       panic(err)
    }

    if cont.Continent == "" {
        fmt.Fprintf(w, "not enough info - json Continent:" + cont.Continent + "\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
	return errors.New("Not enough info")
    }
    return err
}

func parseCountryJson(w http.ResponseWriter, r *http.Request, country *Country) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
    }
    err = json.Unmarshal(body, country)
    if err != nil {
        fmt.Fprint(w, err.Error())
       panic(err)
    }

    if country.Country == "" {
        fmt.Fprintf(w, "not enough info - json Country:" + country.Country + "\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
	return errors.New("Not enough info")
    }
    return err
}

func parseCityJson(w http.ResponseWriter, r *http.Request, city *City) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
	return err
    }
    err = json.Unmarshal(body, city)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
	return err
    }
    if city.City == "" {
        fmt.Fprintf(w, "not enough info - json City:" + city.City + "\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
	return errors.New("Not enough info")
    }
    return err
}

func parseChangeNameJson(w http.ResponseWriter, r *http.Request, names *ChangeName) error {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
    }
    err = json.Unmarshal(body, &names)
    if err != nil {
        fmt.Fprint(w, err.Error())
        panic(err)
	return err
    }
    if names.OldName == "" || names.NewName == "" {
	    fmt.Fprintf(w, "not enough info - json OldName: " + names.OldName + ", NewName: " + names.NewName + "\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
	return errors.New("Not enough info")
    }
    return err
}

func readContPage(w http.ResponseWriter, r *http.Request) {
    if len(keyContinentValId) != 0 {
        fmt.Fprintf(w, "list of continents in DB in alphabetical order\n")
        conts := make([]string, 0, len(keyContinentValId))
        for cont := range keyContinentValId {
            conts = append(conts, cont)
        }
        sort.Strings(conts)
        for _, cont := range conts {
            fmt.Fprintf(w, "   " + cont + "\n")
        }
    } else {
	fmt.Fprintf(w, "no continents in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }

}

func readCountryPage(w http.ResponseWriter, r *http.Request) {
    if len(keyCountryValId) != 0 {
        fmt.Fprintf(w, "list of countries in DB in alphabetical order\n")
        countries := make([]string, 0, len(keyCountryValId))
        for country := range keyCountryValId {
            countries = append(countries, country)
        }
        sort.Strings(countries)
        for _, country := range countries {
            fmt.Fprintf(w, "   " + country + "\n")
        }
    } else {
	fmt.Fprintf(w, "no countries in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func readCityPage(w http.ResponseWriter, r *http.Request) {
    if len(keyCityValId) != 0 {
        fmt.Fprintf(w, "list of cities in DB in alphabetical order\n")
        cities := make([]string, 0, len(keyCityValId))
        for city := range keyCityValId {
        cities = append(cities, city)
        }
        sort.Strings(cities)
        for _, city := range cities {
            fmt.Fprintf(w, "   " + city + "\n")
        }
    } else {
	fmt.Fprintf(w, "no cities in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func readCountryInfoPage(w http.ResponseWriter, r *http.Request) {
    var country Country
    err := parseCountryJson(w, r, &country)
    if err != nil {
        return
    }
    if countryRecord, ok := keyCountryValId[country.Country]; ok {
        fmt.Fprintf(w, "info of country " + country.Country + " in DB\n")
	fmt.Fprintf(w, "   continent: " + *countryRecord.Continent + "\n")
    } else {
	fmt.Fprintf(w, "country " + country.Country + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func readCityInfoPage(w http.ResponseWriter, r *http.Request) {
    var city City
    err := parseCityJson(w, r, &city)
    if err != nil {
        return
    }
    if cityRecord, ok := keyCityValId[city.City]; ok {
        fmt.Fprintf(w, "info of city " + city.City + " in DB\n")
	fmt.Fprintf(w, "   continent: " + *cityRecord.Continent + "\n")
	fmt.Fprintf(w, "   country: " + *cityRecord.Country + "\n")
    } else {
	fmt.Fprintf(w, "city " + city.City + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func readContCountryPage(w http.ResponseWriter, r *http.Request) {
    var cont Continent
    err := parseContJson(w, r, &cont)
    if err != nil {
        return
    }
    if contRecord, ok := keyContinentValId[cont.Continent]; ok {
        fmt.Fprintf(w, "list of countries on " + cont.Continent + " in DB in alphabetical order\n")
	var ptrList []*string
	ptrList = keyContValCountries[contRecord.Id]
        countries := make([]string, 0, len(ptrList))
        for _, ptr := range ptrList {
	    countries = append(countries, *ptr)
        }
        sort.Strings(countries)
        for _, country := range countries {
	    fmt.Fprintf(w, "   " + country + "\n")
        }
    } else {
	fmt.Fprintf(w, "continent " + cont.Continent + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }

}

func readContCountryCityPage(w http.ResponseWriter, r *http.Request) {
    var country ContCountry
    err := parseContCountryJson(w, r, &country)
    if err != nil {
        return
    }
    contRecord := keyContinentValId[country.Continent]
    countryRecord := keyCountryValId[country.Country]
    if contRecord.Id != 0 && countryRecord.Id != 0 {
        key := fmt.Sprintf("%d.%d", contRecord.Id, countryRecord.Id)
        fmt.Fprintf(w, "list of cities of " + country.Country + " on " + country.Continent + " in DB in alphabetical order\n")
	var ptrList []*string
	ptrList = keyContCountryValCities[key]
        cities := make([]string, 0, len(ptrList))
        for _, ptr := range ptrList {
	    cities = append(cities, *ptr)
        }
        sort.Strings(cities)
        for _, city := range cities {
	    fmt.Fprintf(w, "   " + city + "\n")
        }
    } else {
        if contRecord.Id == 0 {
	    fmt.Fprint(w, "continent " + country.Continent + " not in DB (use POST to create)\n")
	}
        if countryRecord.Id == 0 {
	    fmt.Fprint(w, "country " + country.Country + " not in DB (use POST to create)\n")
	}
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func readContCityPage(w http.ResponseWriter, r *http.Request) {
    var cont Continent
    err := parseContJson(w, r, &cont)
    if err != nil {
        return
    }
    if contRecord, ok := keyContinentValId[cont.Continent]; ok {
        fmt.Fprintf(w, "list of cities on " + cont.Continent + " in DB in alphabetical order\n")
	var ptrList []*string
	ptrList = keyContValCities[contRecord.Id]
        cities := make([]string, 0, len(ptrList))
        for _, ptr := range ptrList {
	    cities = append(cities, *ptr)
        }
        sort.Strings(cities)
        for _, city := range cities {
	    fmt.Fprintf(w, "   " + city + "\n")
        }
    } else {
	fmt.Fprintf(w, "continent " + cont.Continent + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func readCountryCityPage(w http.ResponseWriter, r *http.Request) {
    var country Country
    err := parseCountryJson(w, r, &country)
    if err != nil {
        return
    }
    if countryRecord, ok := keyCountryValId[country.Country]; ok {
        fmt.Fprintf(w, "list of cities on " + country.Country + " in DB in alphabetical order\n")
	var ptrList []*string
	ptrList = keyCountryValCities[countryRecord.Id]
        cities := make([]string, 0, len(ptrList))
        for _, ptr := range ptrList {
	    cities = append(cities, *ptr)
        }
        sort.Strings(cities)
        for _, city := range cities {
	    fmt.Fprintf(w, "   " + city + "\n")
        }
    } else {
	fmt.Fprintf(w, "country " + country.Country + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func findContinentEntry(w http.ResponseWriter, key string) *string {
    lowExpKey := strings.ToLower(key)
    for mKey := range keyContinentValId {
        lowKey := strings.ToLower(mKey)
	if lowExpKey == lowKey {
	    return keyContinentValId[mKey].NamePtr
	}
    }
    return nil
}

func findCountryEntry(w http.ResponseWriter, key string) *string {
    lowExpKey := strings.ToLower(key)
    for mKey := range keyCountryValId {
        lowKey := strings.ToLower(mKey)
	if lowExpKey == lowKey {
	    return keyCountryValId[mKey].NamePtr
	}
    }
    return nil
}

func findCityEntry(w http.ResponseWriter, key string) *string {
    lowExpKey := strings.ToLower(key)
    for mKey := range keyCityValId {
        lowKey := strings.ToLower(mKey)
	if lowExpKey == lowKey {
            return keyCityValId[mKey].NamePtr
	}
    }
    return nil
}

func createContCountryMapEntry(w http.ResponseWriter, cont string, cPtr *string) {
    if contRecord, ok := keyContinentValId[cont]; ok {
	ptrList := keyContValCountries[contRecord.Id]
	for _, ptr := range ptrList {
            if *ptr == *cPtr {
                return
            }
        }
	ptrList = append(ptrList, cPtr)
        mutex.Lock()
        keyContValCountries[contRecord.Id] = ptrList
        mutex.Unlock()
    } else {
	fmt.Fprintf(w, "continent " + cont + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func deleteContCountryMapEntry(w http.ResponseWriter, contId int, cPtr *string) {
    var newPtrList []*string
    ptrList := keyContValCountries[contId]
    for _, ptr := range ptrList {
        if *ptr != *cPtr {
            newPtrList = append(newPtrList, ptr)
        }
    }
    mutex.Lock()
    keyContValCountries[contId] = newPtrList
    mutex.Unlock()
}

func createContCountryCityMapEntry(w http.ResponseWriter, cont string, country string, cPtr *string) {
    contRecord := keyContinentValId[cont]
    countryRecord := keyCountryValId[country]
    if contRecord.Id != 0 && countryRecord.Id != 0 {
        key := fmt.Sprintf("%d.%d", contRecord.Id, countryRecord.Id)
        ptrList := keyContCountryValCities[key]
        for _, ptr := range ptrList {
            if *ptr == *cPtr {
                return
            }
        }
	ptrList = append(ptrList, cPtr)
        mutex.Lock()
        keyContCountryValCities[key] = ptrList
        mutex.Unlock()
    } else { //records should always exist, since created upon create/update if not exist
        if contRecord.Id == 0 {
	    fmt.Fprint(w, "continent " + cont + " not in DB (use POST to create)\n")
	}
        if countryRecord.Id == 0 {
	    fmt.Fprint(w, "country " + country + " not in DB (use POST to create)\n")
	}
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func deleteContCountryCityMapEntry(w http.ResponseWriter, contId int, countryId int, cPtr *string, remCityOnly bool) {
    key := fmt.Sprintf("%d.%d", contId, countryId)
    if remCityOnly {
        var newPtrList []*string
        ptrList := keyContCountryValCities[key]
        for _, ptr := range ptrList {
            if *ptr != *cPtr {
                newPtrList = append(newPtrList, ptr)
            }
        }
        mutex.Lock()
        keyContCountryValCities[key] = newPtrList
        mutex.Unlock()
    } else {
	delete(keyContCountryValCities, key)
    }
}

func createContCityMapEntry(w http.ResponseWriter, cont string, cPtr *string) {
    if contRecord, ok := keyContinentValId[cont]; ok {
	ptrList := keyContValCities[contRecord.Id]
	for _, ptr := range ptrList {
            if *ptr == *cPtr {
                return
            }
        }
	ptrList = append(ptrList, cPtr)
        mutex.Lock()
        keyContValCities[contRecord.Id] = ptrList
        mutex.Unlock()
    } else {
	fmt.Fprintf(w, "continent " + cont + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func deleteContCityMapEntry(w http.ResponseWriter, contId int, cPtr *string) {
    var newPtrList []*string
    ptrList := keyContValCities[contId]
    for _, ptr := range ptrList {
        if *ptr != *cPtr {
            newPtrList = append(newPtrList, ptr)
        }
    }
    mutex.Lock()
    keyContValCities[contId] = newPtrList
    mutex.Unlock()
}

func createCountryCityMapEntry(w http.ResponseWriter, country string, cPtr *string) {
    if countryRecord, ok := keyCountryValId[country]; ok {
	ptrList := keyCountryValCities[countryRecord.Id]
	for _, ptr := range ptrList {
            if *ptr == *cPtr {
                return
            }
        }
	ptrList = append(ptrList, cPtr)
        mutex.Lock()
        keyCountryValCities[countryRecord.Id] = ptrList
        mutex.Unlock()
    } else {
	fmt.Fprintf(w, "country " + country + " not in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func deleteCountryCityMapEntry(w http.ResponseWriter, countryId int, cPtr *string) {
    var newPtrList []*string
    ptrList := keyCountryValCities[countryId]
    for _, ptr := range ptrList {
        if *ptr != *cPtr {
            newPtrList = append(newPtrList, ptr)
        }
    }
    mutex.Lock()
    keyCountryValCities[countryId] = newPtrList
    mutex.Unlock()
}

func deleteCitiesOfCountryCityMapEntry(w http.ResponseWriter, contId int, countryId int) {
    ptrList := keyCountryValCities[countryId]
    for _, ptr := range ptrList {
        //delete city from keyContValCity map
        deleteContCityMapEntry(w, contId, ptr)
	//delete city DB entry
	deleteCityDbEntry(*ptr)
    }
    //delete country from keyCountryValCities
    delete(keyCountryValCities, countryId)
}

func deleteCountriesAndCitiesOfContCountryMapEntry(w http.ResponseWriter, contId int) {
    //delete countries
    ptrList := keyContValCountries[contId]
    for _, ptr := range ptrList {
        countryRecord := keyCountryValId[*ptr]
        //delete entry from search map trees
        deleteContCountryCityMapEntry(w, contId, countryRecord.Id, nil, false)
	delete(keyContValCountries, contId)
	delete(keyCountryValCities, countryRecord.Id)
	//delete country DB entry
	deleteCountryDbEntryWithId(*ptr, countryRecord.Id)
    }
    //delete cities
    ptrList = keyContValCities[contId]
    for _, ptr := range ptrList {
        //delete city DB entry
	deleteCityDbEntry(*ptr)
    }
    delete(keyContValCities, contId)
}

func createContinentMapEntry(w http.ResponseWriter, key string) *string {
    id := getNewId("continent")
    var continent ContinentRecord
    continent.Id = id
    continent.Name = key
    continent.NamePtr = &continent.Name
    mutex.Lock()
    keyContinentValId[key] = continent
    mutex.Unlock()
    fmt.Fprintf(w, "created continent " + key + " with ID " + strconv.Itoa(id) +  "\n")
    return continent.NamePtr
}

func createContinent(w http.ResponseWriter, r *http.Request) {
    var cont Continent
    err := parseContJson(w, r, &cont)
    if err != nil {
        return
    }
    cPtr := findContinentEntry(w, cont.Continent)
    if cPtr == nil {
	//create DB entry
        createContinentMapEntry(w, cont.Continent)
    } else {
	fmt.Fprintf(w, "continent " + cont.Continent + " already exists in DB (use PUT to update)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}
func updateCountryMapEntry(w http.ResponseWriter, key string, cont *string) *string {
    country := keyCountryValId[key]
    oldContinent := country.Continent
    if *country.Continent != *cont {
        country.Continent = cont
	country.NamePtr = &country.Name
        mutex.Lock()
        keyCountryValId[key] = country
        mutex.Unlock()
        fmt.Fprintf(w, "updated country " + key + " with ID " + strconv.Itoa(country.Id) + ", cont: " + *cont + ", old cont: " + *oldContinent + "\n")
    }
    return country.NamePtr
}


func createCountryMapEntry(w http.ResponseWriter, key string, cont *string) *string {
    id := getNewId("country")
    var country CountryRecord
    country.Id = id
    country.Name = key
    country.NamePtr = &country.Name
    country.Continent = cont
    mutex.Lock()
    keyCountryValId[key] = country
    mutex.Unlock()
    fmt.Fprintf(w, "created country " + key + " with ID " + strconv.Itoa(id) + " cont: " + *cont +  "\n")
    return country.NamePtr
}

func createCountry(w http.ResponseWriter, r *http.Request) {
    var country ContCountry
    err := parseContCountryJson(w, r, &country)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCountryEntry(w, country.Country)
    if cPtr == nil {
        var contPtr *string
	contPtr = findContinentEntry(w, country.Continent)
        if contPtr == nil {
	    contPtr = createContinentMapEntry(w, country.Continent)
	}
	//create DB entry
        cPtr = createCountryMapEntry(w, country.Country, contPtr)
	//add to search map
	createContCountryMapEntry(w, country.Continent, cPtr)
    } else {
	fmt.Fprintf(w, "country " + country.Country + " already exists in DB (use PUT to update)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func createCityMapEntry(w http.ResponseWriter, key string, country *string, cont *string) *string {
    id := getNewId("city")
    var city CityRecord
    city.Id = id
    city.Name = key
    city.NamePtr = &city.Name
    city.Country = country
    city.Continent = cont
    mutex.Lock()
    keyCityValId[key] = city
    mutex.Unlock()
    fmt.Fprintf(w, "created city " + key + " with ID " + strconv.Itoa(id) + " country: " + *country + ", cont: " + *cont +  "\n")
    return city.NamePtr
}

func updateCityMapEntry(w http.ResponseWriter, key string, country *string, cont *string) (*string, *string, *string) {
    city := keyCityValId[key]
    oldCountry := city.Country
    oldContinent := city.Continent
    if *city.Continent != *cont || *city.Country != *country {
        city.NamePtr = &city.Name
        if *city.Country != *country {
            city.Country = country
	}
	if *city.Continent != *cont {
            city.Continent = cont
	}
        mutex.Lock()
        keyCityValId[key] = city
        mutex.Unlock()
        fmt.Fprintf(w, "updated city " + key + " with ID " + strconv.Itoa(city.Id) + " country: " + *country + ", cont: " + *cont + ", old country: " + *oldCountry + ", old cont: " + *oldContinent + "\n")
    }
    return city.NamePtr, oldCountry, oldContinent
}

func createCity(w http.ResponseWriter, r *http.Request) {
    var city ContCountryCity
    err := parseContCountryCityJson(w, r, &city)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCityEntry(w, city.City)
    if cPtr == nil {
        var contPtr *string
	contPtr = findContinentEntry(w, city.Continent)
        if contPtr == nil {
	    contPtr = createContinentMapEntry(w, city.Continent)
	}
	var countryPtr *string
	countryPtr = findCountryEntry(w, city.Country)
	if countryPtr == nil {
	    countryPtr = createCountryMapEntry(w, city.Country, contPtr)
	}
	//create DB entry
        cPtr = createCityMapEntry(w, city.City, countryPtr, contPtr)
	//add to search map
	createContCountryCityMapEntry(w, city.Continent, city.Country, cPtr)
	createContCountryMapEntry(w, city.Continent, countryPtr)
	createContCityMapEntry(w, city.Continent, cPtr)
	createCountryCityMapEntry(w, city.Country, cPtr)
    } else {
	fmt.Fprintf(w, "city " + city.City + " already exists in DB (use PUT to update)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func updateCityMapKeyAndName(cPtr *string, cityRecord *CityRecord, names ChangeName) {
    cityRecord.Name = names.NewName
    cityRecord.NamePtr = &cityRecord.Name
    keyCityValId[names.NewName] = *cityRecord
    delete(keyCityValId, *cPtr)
}

func updateCityName(w http.ResponseWriter, r *http.Request) {
    var names ChangeName
    err := parseChangeNameJson(w, r, &names)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCityEntry(w, names.OldName)
    if cPtr != nil {
	//update DB entry
	cityRecord := keyCityValId[*cPtr]
        updateCityMapKeyAndName(cPtr, &cityRecord, names)
        //add with new info to search map and delete with old info from search map
        contRecord := keyContinentValId[*cityRecord.Continent]
	countryRecord := keyCountryValId[*cityRecord.Country]
        createContCountryCityMapEntry(w, contRecord.Name, countryRecord.Name, cityRecord.NamePtr)
        deleteContCountryCityMapEntry(w, contRecord.Id, countryRecord.Id, cPtr, true)
        createContCityMapEntry(w, contRecord.Name, cityRecord.NamePtr)
        deleteContCityMapEntry(w, contRecord.Id, cPtr)
        createCountryCityMapEntry(w, countryRecord.Name, cityRecord.NamePtr)
        deleteCountryCityMapEntry(w, countryRecord.Id, cPtr)
	fmt.Fprintf(w, "updated city name to: " + names.NewName + " - old name: " + names.OldName + "\n")
    } else {
	fmt.Fprintf(w, "city " + names.OldName + " does not exists in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func updateCity(w http.ResponseWriter, r *http.Request) {
    var city ContCountryCity
    err := parseContCountryCityJson(w, r, &city)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCityEntry(w, city.City)
    if cPtr != nil {
        var contPtr *string
	contPtr = findContinentEntry(w, city.Continent)
        if contPtr == nil {
	    contPtr = createContinentMapEntry(w, city.Continent)
	}
	var countryPtr *string
	countryPtr = findCountryEntry(w, city.Country)
	if countryPtr == nil {
	    countryPtr = createCountryMapEntry(w, city.Country, contPtr)
	}
	//update DB entry
	cityRecord := keyCityValId[*cPtr]
        if contPtr != nil && countryPtr != nil &&
	   (*cityRecord.Continent != *contPtr || *cityRecord.Country != *countryPtr) {
            contRecord := keyContinentValId[*cityRecord.Continent]
            countryRecord := keyCountryValId[*cityRecord.Country]
            updateCityMapEntry(w, *cPtr, countryPtr, contPtr)
            //add with new info to search map and delete with old info from search map
	    createContCountryCityMapEntry(w, *contPtr, *countryPtr, cPtr)
            deleteContCountryCityMapEntry(w, contRecord.Id, countryRecord.Id, cPtr, true)
	    if *cityRecord.Continent != *contPtr {
                updateCountryMapEntry(w, *countryPtr, contPtr)
                createContCountryMapEntry(w, *contPtr, countryPtr)
                deleteContCountryMapEntry(w, contRecord.Id, countryRecord.NamePtr)
                createContCityMapEntry(w, *contPtr, cPtr)
                deleteContCityMapEntry(w, contRecord.Id, cPtr)
	    }
	    if *cityRecord.Country != *countryPtr {
                createContCountryMapEntry(w, *contPtr, countryPtr)
	        createCountryCityMapEntry(w, *countryPtr, cPtr)
                deleteCountryCityMapEntry(w, countryRecord.Id, cPtr)
	    }
	}
    } else {
	fmt.Fprintf(w, "city " + city.City + " does not exists in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func updateCountryMapKeyAndName(cPtr *string, countryRecord *CountryRecord, names ChangeName) {
    countryRecord.Name = names.NewName
    countryRecord.NamePtr = &countryRecord.Name
    keyCountryValId[names.NewName] = *countryRecord
    delete(keyCountryValId, *cPtr)
}

func updateCountryName(w http.ResponseWriter, r *http.Request) {
    var names ChangeName
    err := parseChangeNameJson(w, r, &names)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCountryEntry(w, names.OldName)
    if cPtr != nil {
	//update DB entry
        countryRecord := keyCountryValId[*cPtr]
        contRecord := keyContinentValId[*countryRecord.Continent]
        updateCountryMapKeyAndName(cPtr, &countryRecord, names)
	//country needed to be updated to cities
        ptrList := keyCountryValCities[countryRecord.Id]
        for _, cityPtr := range ptrList {
            updateCityMapEntry(w, *cityPtr, countryRecord.NamePtr, &contRecord.Name)
        }
        //add with new info to search map and delete with old info from search map
        createContCountryMapEntry(w, contRecord.Name, countryRecord.NamePtr)
        deleteContCountryMapEntry(w, contRecord.Id, cPtr)
	fmt.Fprintf(w, "updated country name to: " + names.NewName + " - old name: " + names.OldName + "\n")
    } else {
	fmt.Fprintf(w, "country " + names.OldName + " does not exists in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func updateCountry(w http.ResponseWriter, r *http.Request) {
    var country ContCountry
    err := parseContCountryJson(w, r, &country)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCountryEntry(w, country.Country)
    if cPtr != nil {
        var contPtr *string
	contPtr = findContinentEntry(w, country.Continent)
        if contPtr == nil {
	    contPtr = createContinentMapEntry(w, country.Continent)
	}
	//update DB entry
	countryRecord := keyCountryValId[*cPtr]
        if contPtr != nil && *countryRecord.Continent != *contPtr {
            contRecord := keyContinentValId[*countryRecord.Continent]
            updateCountryMapEntry(w, *cPtr, contPtr)
            //add with new info to search map and delete with old info from search map
	    //before updating keyContCountryValCities, needed updates done to cities
            key := fmt.Sprintf("%d.%d", contRecord.Id, countryRecord.Id)
	    ptrList := keyContCountryValCities[key]
	    for _, cityPtr := range ptrList {
                updateCityMapEntry(w, *cityPtr, cPtr, contPtr)
                createContCityMapEntry(w, *contPtr, cityPtr)
                deleteContCityMapEntry(w, contRecord.Id, cityPtr)
	        createContCountryCityMapEntry(w, *contPtr, *cPtr, cityPtr)
                deleteContCountryCityMapEntry(w, contRecord.Id, countryRecord.Id, cityPtr, true)
            }
            createContCountryMapEntry(w, *contPtr, cPtr)
            deleteContCountryMapEntry(w, contRecord.Id, cPtr)
	}
    } else {
	fmt.Fprintf(w, "country " + country.Country + " does not exists in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func changeContMapKey(origMap map[string]ContinentRecord, names ChangeName) map[string]ContinentRecord {
    newMap := make(map[string]ContinentRecord)
    for key, value := range origMap {
        if key == names.OldName {
            newMap[names.NewName] = value
        } else {
            newMap[key] = value
        }
    }
    return newMap
}

func updateContMapKeyAndName(cPtr *string, contRecord *ContinentRecord, names ChangeName) {
    contRecord.Name = names.NewName
    contRecord.NamePtr = &contRecord.Name
    keyContinentValId[names.NewName] = *contRecord
    delete(keyContinentValId, *cPtr)
}

func updateContinentName(w http.ResponseWriter, r *http.Request) {
    var names ChangeName
    err := parseChangeNameJson(w, r, &names)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findContinentEntry(w, names.OldName)
    if cPtr != nil {
	//update DB entry
        contRecord := keyContinentValId[*cPtr]
        updateContMapKeyAndName(cPtr, &contRecord, names)
	//contRecord := keyContinentValId[*cPtr]
	//contRecord.Name = names.NewName
	//keyContinentValId[*cPtr] = contRecord
        //keyContinentValId = changeContMapKey(keyContinentValId, names)
	fmt.Fprintf(w, "updated continent name to: " + names.NewName + " - old name: " + names.OldName + "\n")
    } else {
	fmt.Fprintf(w, "continent " + names.OldName + " does not exists in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
    fmt.Println("keyContCountryValCities")
    fmt.Println(keyContCountryValCities)
    fmt.Println("keyContValCountries")
    fmt.Println(keyContValCountries)
    fmt.Println("keyContValCities")
    fmt.Println(keyContValCities)
    fmt.Println("keyCountryValCities")
    fmt.Println(keyCountryValCities)
    fmt.Println("keyContinentValId")
    fmt.Println(keyContinentValId)
    fmt.Println("keyCountryValId")
    fmt.Println(keyCountryValId)
    fmt.Println("keyCityValId")
    fmt.Println(keyCityValId)
}

func deleteCity(w http.ResponseWriter, r *http.Request) {
    var city City
    err := parseCityJson(w, r, &city)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCityEntry(w, city.City)
    if cPtr != nil {
        cityRecord := keyCityValId[*cPtr]
	// cityRecord pointers should never be nil, since required info on 
	// create and update and the whole record removed upon delete 
	if cityRecord.Continent != nil && cityRecord.Country != nil {
	    contRecord := keyContinentValId[*cityRecord.Continent]
	    countryRecord := keyCountryValId[*cityRecord.Country]
	    //delete from search map
            deleteContCountryCityMapEntry(w, contRecord.Id, countryRecord.Id, cPtr, true)
	    deleteContCityMapEntry(w, contRecord.Id, cPtr)
	    deleteCountryCityMapEntry(w, countryRecord.Id, cPtr)
	} else {
            fmt.Fprintf(w, "city " + *cPtr + " entry corrupted in DB (contPtr: " + *cityRecord.Continent + " countryPtr: " + *cityRecord.Country + ")\n")
	}
	deleteCityDbEntryWithId(*cPtr, cityRecord.Id)
        fmt.Fprintf(w, "city " + *cPtr + " entry deleted from DB (cont: " + *cityRecord.Continent + " country: " + *cityRecord.Country + ")\n")
    } else {
	fmt.Fprintf(w, "city " + city.City + " does not exist in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func deleteCountry(w http.ResponseWriter, r *http.Request) {
    var country Country
    err := parseCountryJson(w, r, &country)
    if err != nil {
        return
    }
    var cPtr *string
    cPtr = findCountryEntry(w, country.Country)
    if cPtr != nil {
        countryRecord := keyCountryValId[*cPtr]
	// countryRecord pointer should never be nil, since required info on 
	// create and update and the whole record removed upon delete 
	if countryRecord.Continent != nil {
	    contRecord := keyContinentValId[*countryRecord.Continent]
	    //delete from search map
	    deleteContCountryMapEntry(w, contRecord.Id, cPtr)
            deleteContCountryCityMapEntry(w, contRecord.Id, countryRecord.Id, nil, false)
	    // delete all city DB entries of the country - otherwise corrupted
	    deleteCitiesOfCountryCityMapEntry(w, contRecord.Id, countryRecord.Id)
	} else {
            fmt.Fprintf(w, "country " + *cPtr + " entry corrupted in DB (contPtr: " + *countryRecord.Continent + ")\n")
	}
	//delete DB entry
	deleteCountryDbEntryWithId(*cPtr, countryRecord.Id)
        fmt.Fprintf(w, "country " + *cPtr + " entry deleted from DB (cont: " + *countryRecord.Continent + ")\n")
    } else {
	fmt.Fprintf(w, "country " + country.Country + " does not exist in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func deleteContinent(w http.ResponseWriter, r *http.Request) {
    var cont Continent
    err := parseContJson(w, r, &cont)
    if err != nil {
        return
    }
    cPtr := findContinentEntry(w, cont.Continent)
    if cPtr != nil {
        contRecord := keyContinentValId[*cPtr]
        //delete from search map
        //delete all country and city DB entries of the cont - otherwise corrupted
        deleteCountriesAndCitiesOfContCountryMapEntry(w, contRecord.Id)
        // and then delete the search map entry
	deleteContDbEntryWithId(*cPtr, contRecord.Id)
        fmt.Fprintf(w, "continent " + *cPtr + " entry deleted from DB\n")
    } else {
	fmt.Fprintf(w, "continent " + cont.Continent + " does not exist in DB (use POST to create)\n")
        fmt.Fprintf(w, "For help: localhost:10000/help\n")
    }
}

func handleRequests() {
    r := mux.NewRouter()
    //Read records
    GETsr := r.Methods("GET").Subrouter()
    //Create records
    POSTsr := r.Methods("POST").Subrouter()
    //Delete records
    DELsr := r.Methods("DELETE").Subrouter()
    //Update records
    PUTsr := r.Methods("PUT").Subrouter()
    r.HandleFunc("/", homePage)
    r.HandleFunc("/help", helpPage)
    GETsr.HandleFunc("/continents", readContPage)
    GETsr.HandleFunc("/countries", readCountryPage)
    GETsr.HandleFunc("/cities", readCityPage)
    GETsr.HandleFunc("/country/info", readCountryInfoPage)
    GETsr.HandleFunc("/city/info", readCityInfoPage)
    GETsr.HandleFunc("/continent/country/cities", readContCountryCityPage)
    GETsr.HandleFunc("/continent/countries", readContCountryPage)
    GETsr.HandleFunc("/continent/cities", readContCityPage)
    GETsr.HandleFunc("/country/cities", readCountryCityPage)
    POSTsr.HandleFunc("/city", createCity)
    POSTsr.HandleFunc("/country", createCountry)
    POSTsr.HandleFunc("/continent", createContinent)
    DELsr.HandleFunc("/city", deleteCity)
    DELsr.HandleFunc("/country", deleteCountry)
    DELsr.HandleFunc("/continent", deleteContinent)
    PUTsr.HandleFunc("/city/name", updateCityName)
    PUTsr.HandleFunc("/country/name", updateCountryName)
    PUTsr.HandleFunc("/continent/name", updateContinentName)
    PUTsr.HandleFunc("/city", updateCity)
    PUTsr.HandleFunc("/country", updateCountry)
    //did not find any elegant way to catch all the rest URLs in go
    r.HandleFunc("/{param1}", helpPage)
    r.HandleFunc("/{param1}/{param2}", helpPage)
    r.HandleFunc("/{param1}/{param2}/{param3}", helpPage)
    r.HandleFunc("/{param1}/{param2}/{param3}/{param4}", helpPage)
    r.HandleFunc("/{param1}/{param2}/{param3}/{param4}/{param5}", helpPage)
    r.HandleFunc("/{param1}/{param2}/{param3}/{param4}/{param5}/{param6}", helpPage)
    log.Fatal(http.ListenAndServe(":10000", r))
}

func main() {
    keyCityValId = make(map[string]CityRecord)
    keyCountryValId = make(map[string]CountryRecord)
    keyContinentValId = make(map[string]ContinentRecord)

    keyContCountryValCities = make(map[string][]*string)
    keyContValCities = make(map[int][]*string)
    keyCountryValCities = make(map[int][]*string)
    keyContValCountries = make(map[int][]*string)

    handleRequests()
}
