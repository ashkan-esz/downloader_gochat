package geoip

import (
	errorHandler "downloader_gochat/pkg/error"
	_ "embed"
	"fmt"
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

var db *maxminddb.Reader

type record struct {
	Country struct {
		Names             map[string]string `maxminddb:"names"`
		IsoCode           string            `maxminddb:"iso_code"`
		GeoNameID         uint              `maxminddb:"geoname_id"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
	} `maxminddb:"country"`
	City struct {
		Names     map[string]string `maxminddb:"names"`
		GeoNameID uint              `maxminddb:"geoname_id"`
	} `maxminddb:"city"`
}

//go:embed GeoLite2-City.mmdb
var geoLiteDb []byte

func Load() {
	//f0, err := os.Getwd()
	//db2, err := maxminddb.Open(f0 + "/pkg/geoip/GeoLite2-City.mmdb")
	db2, err := maxminddb.FromBytes(geoLiteDb)
	if err != nil {
		log.Fatal(err)
	}
	db = db2
}

func GetRequestLocation(ipString string) string {
	var r record
	ip := net.ParseIP(ipString)
	err := db.Lookup(ip, &r)
	if err != nil {
		errorMessage := fmt.Sprintf("error on lookup ip location: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return ""
	}
	country := r.Country.Names["en"]
	if country == "" {
		return ""
	}
	city := r.City.Names["en"]
	if city == "" {
		return country
	}
	return city + ", " + country
}
