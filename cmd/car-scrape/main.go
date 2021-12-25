package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http/pprof"

	"github.com/PuerkitoBio/goquery"
)

type Car struct {
	ID          uint   `gorm:"primarykey"`
	AdID        string `gorm:"unique"`
	Title       string `json:",omitempty"`
	Stars       string
	Noticed     string
	Updated     string
	Price       string `json:",omitempty"`
	DataPrice   string
	ContactName string
	Location    string
	Phone       string
	DataPhone   string
	Description string
	PhotoLink   string
	PhotoLinks  string
	Link        string
	SeenAt      *time.Time
	UpdatedAt   *time.Time
	TakenOutAt  *time.Time
	Sold        bool
	SoldIn      string
	ErrorMsg    string
	Changes     string

	// Params
	Eksportui           string
	PagaminimoData      string
	Rida                string
	Variklis            string
	KuroTipas           string
	KebuloTipas         string
	DuruSkaicius        string
	VarantiejiRatai     string
	PavaruDeze          string
	KlimatoValdymas     string
	Spalva              string
	VairoPadetis        string
	TechApziuraIki      string
	RatlankiuSkersmuo   string
	NuosavaMase         string
	SedimuVietuSkaicius string
	KebuloNumerisVIN    string
	Emisija             string
	TarsosMokestis      string
	PirmosRegSalis      string
	SDK                 string
	EuroStandartas      string
	Mieste              string
	Uzmiestyje          string
	Vidutines           string
	Leftover            string

	// Features
	Salonas      string
	Elektronika  string
	Apsauga      string
	Audio        string
	Eksterjeras  string
	Kiti         string
	Saugumas     string
	FeaturesLeft string
}

var DB *Database

func ExampleScrape() {
	var err error
	DB, err = Connect()
	if err != nil {
		fmt.Println(err)
	}
	DB.DB.AutoMigrate(&Car{})
	BrutalScrape()
	//OneTimeScan()
}

type CarMini struct {
	Link string
	AdID string
}

func BrutalScrape() {
	link := os.Getenv("CAR_URL") // should consist of price var low %d, high %d, page var number %d,
	for price := 0; price < 100000; price += 1000 {
		low := price
		high := price + 1000 - 1
		page := OpenPage(fmt.Sprintf(link, low, high, 1))
		pages := getPageCount(page)
		for i := 1; i <= pages; i++ {
			page := OpenPage(fmt.Sprintf(link, low, high, i))
			carMinis := ReadCarList(page)
			for _, mini := range carMinis {
				if mini.AdID == "" {
					fmt.Println("id empty")
					continue
				}
				if exists(mini.AdID) {
					fmt.Println("exists ", mini.AdID)
					continue
				}
				carPage := OpenPage(mini.Link)
				car := ReadCar(carPage)
				create(car)
			}
		}
	}
}

func OneTimeScan() {
	//re := OpenFile("_data/cars.html")
	//ReadCarList(re)
	cr := OpenFile("_data/car.html")
	car := ReadCar(cr)
	create(car)
	/*err := DB.DB.Create(car).Error
	if err != nil {
		fmt.Println(err)
	}*/

	// tests updates only non zero fields
	/*car2 := &Car{ID: car.ID, Title: "New title", DataPrice: "123456789"}
	err = DB.DB.Updates(car2).Error
	if err != nil {
		fmt.Println(err)
	}*/
}

func OpenFile(filename string) io.Reader {
	re, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	return re
}

func OpenPage(link string) io.Reader {
	for {
		// Request the HTML page.
		res, err := http.Get(link)
		if err != nil {
			log.Fatal(err)
		}
		//defer res.Body.Close()
		if res.StatusCode == 429 {
			fmt.Println("received 429. Retrying")
			time.Sleep(time.Second)
		} else if res.StatusCode == 200 {
			return res.Body
		} else {
			log.Println(link)
			log.Fatalf("other status code error: %d %s", res.StatusCode, res.Status)
		}
	}
}

func ReadCarList(re io.Reader) []CarMini {
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(re)
	if err != nil {
		log.Fatal(err)
	}

	list := doc.Find(".auto-lists")

	minis := make([]CarMini, 0)
	// links := make([]string, 0)
	// Find the review items
	list.Find(".announcement-item").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		link, ok := s.Attr("href")
		if !ok {
			fmt.Println("error. link missing")
			return
		}
		bod := s.Find(".announcement-body")
		_ = bod.Find(".stars-badge").Text() //stars
		anBook := bod.Find(".announcement-bookmark-button")
		dataId, ok := anBook.Attr("data-id")
		if !ok {
			fmt.Print("ad-id not found")
			return
		}

		mini := CarMini{
			Link: link,
			AdID: dataId,
		}
		minis = append(minis, mini)
		//carInfoFromList(s)
	})
	return minis
}

func ReadCar(re io.Reader) *Car {
	doc, err := goquery.NewDocumentFromReader(re)
	if err != nil {
		log.Fatal(err)
	}

	body := doc.Find("body")
	title := strings.TrimSpace(body.Find("h1").Text())
	noticed := body.Find(".bookmark-stats-bar").Find("b").Text()
	updated := body.Find(".bookmark-stats-bar").Find(".bar-item").Last().Text()
	price := strings.TrimSpace(body.Find(".price").Text())
	contactName := strings.TrimSpace(body.Find(".seller-contact-name").Text())
	location := strings.TrimSpace(body.Find(".seller-contact-location").Text())
	numSel := body.Find(".seller-phone-number")
	phone := strings.TrimSpace(numSel.Text())
	dataPrice, _ := numSel.Attr("data-price")
	dataId, _ := body.Find(".action-button-share").Attr("data-id")
	dataPhone, _ := numSel.Attr("data-clipboard-text")
	description := strings.TrimSpace(body.Find(".announcement-description").Text())
	reg, _ := regexp.Compile(`\s?\n+\s+`)
	description = reg.ReplaceAllString(description, "\n")
	link, _ := body.Find(".action-button-copy").Attr("data-clipboard-text")
	anMedia := body.Find(".announcement-media-gallery")
	photoLink, _ := anMedia.Find(".thumbnail").First().Find("img").Attr("src")
	photoLinksSlice := make([]string, 0)
	anMedia.Find(".thumbnail").Each(func(i int, f *goquery.Selection) {
		atr, ok := f.Attr("style")
		if !ok {
			return
		}
		lnk := atr[23 : len(atr)-2]
		photoLinksSlice = append(photoLinksSlice, lnk)
	})
	photoLinks := strings.Join(photoLinksSlice, ",")
	containerSel := body.Find(".content-container")
	soldSel := containerSel.Find(".is-sold").Find(".is-sold-badge")
	soldIn := strings.TrimSpace(soldSel.Find("span").First().Text())
	erMsg := strings.TrimSpace(containerSel.Find(".error").Find(".error-msg").Find(".msg-subject").Text())
	salonas := ""
	elektronika := ""
	apsauga := ""
	audio := ""
	eksterjeras := ""
	kiti := ""
	saugumas := ""
	featuresLeft := make([]string, 0)
	body.Find(".feature-row").Each(func(i int, f *goquery.Selection) {
		features := make([]string, 0)
		f.Find(".feature-list").Find(".feature-item").Each(func(i int, f2 *goquery.Selection) {
			name2 := strings.TrimSpace(f2.Text())
			features = append(features, name2)
		})
		feature := strings.Join(features, ",")
		tt := strings.TrimSpace(f.Find(".feature-label").Text())
		switch tt {
		case "Salonas":
			salonas = feature
		case "Elektronika":
			elektronika = feature
		case "Apsauga":
			apsauga = feature
		case "Audio/video įranga":
			audio = feature
		case "Eksterjeras":
			eksterjeras = feature
		case "Kiti ypatumai":
			kiti = feature
		case "Saugumas":
			saugumas = feature
		default:
			featuresLeft = append(featuresLeft, feature)
		}
	})
	featuresLeftLine := strings.Join(featuresLeft, ";")

	now := time.Now()
	car := &Car{
		AdID:        dataId,
		Title:       title,
		Noticed:     noticed,
		Updated:     updated,
		Price:       price,
		DataPrice:   dataPrice,
		ContactName: contactName,
		Location:    location,
		Phone:       phone,
		DataPhone:   dataPhone,
		Description: description,
		PhotoLink:   photoLink,
		PhotoLinks:  photoLinks,
		Link:        link,
		SeenAt:      &now,
		UpdatedAt:   &now,
		TakenOutAt:  nil,
		Sold:        soldIn != "",
		SoldIn:      soldIn,
		ErrorMsg:    erMsg,
		// Stars:               "",
		Salonas:      salonas,
		Elektronika:  elektronika,
		Apsauga:      apsauga,
		Audio:        audio,
		Eksterjeras:  eksterjeras,
		Kiti:         kiti,
		Saugumas:     saugumas,
		FeaturesLeft: featuresLeftLine,
	}

	// Sets parameters from parameter-row
	body.Find(".parameter-row").Each(func(i int, f *goquery.Selection) {
		label := strings.TrimSpace(f.Find(".parameter-label").Text())
		value := strings.TrimSpace(f.Find(".parameter-value").Text())
		if label != "" && value != "" {
			putParam(car, label, value)
		}
	})

	return car
}

func putParam(car *Car, label string, value string) {
	switch label {
	case "Eksportui":
		car.Eksportui = value
	case "Pagaminimo data":
		car.PagaminimoData = value
	case "Rida":
		car.Rida = value
	case "Variklis":
		car.Variklis = value
	case "Kuro tipas":
		car.KuroTipas = value
	case "Kėbulo tipas":
		car.KebuloTipas = value
	case "Durų skaičius":
		car.DuruSkaicius = value
	case "Varantieji ratai":
		car.VarantiejiRatai = value
	case "Pavarų dėžė":
		car.PavaruDeze = value
	case "Klimato valdymas":
		car.KlimatoValdymas = value
	case "Spalva":
		car.Spalva = value
	case "Vairo padėtis":
		car.VairoPadetis = value
	case "Tech. apžiūra iki":
		car.TechApziuraIki = value
	case "Ratlankių skersmuo":
		car.RatlankiuSkersmuo = value
	case "Nuosava masė, kg":
		car.NuosavaMase = value
	case "Sėdimų vietų skaičius":
		car.SedimuVietuSkaicius = value
	case "Kėbulo numeris (VIN)":
		car.KebuloNumerisVIN = value
	case "Pirmosios registracijos šalis":
		car.PirmosRegSalis = value
	case "SDK":
		car.SDK = value
	case "Euro standartas":
		car.EuroStandartas = value
	case "CO₂ emisija, g/km":
		car.Emisija = value
	case "Taršos mokestis":
		car.TarsosMokestis = value
	case "Mieste":
		car.Mieste = value
	case "Užmiestyje":
		car.Uzmiestyje = value
	case "Vidutinės":
		car.Vidutines = value
	default:
		car.Leftover += fmt.Sprintf("%s:%s;", label, value)
	}
}

func create(car *Car) {
	if exists(car.AdID) {
		fmt.Println("exists", car.AdID)
		return
	}
	// findUpdate(car)
	// Found
	err := DB.DB.Create(car).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	imageCreated := "image"
	fileName := "images/" + car.AdID + ".jpg"
	URL := car.PhotoLink
	err = downloadFile(URL, fileName)
	if err != nil {
		imageCreated = "noimg"
	}
	fmt.Println("created", car.AdID, imageCreated)
}

func exists(adID string) bool {
	found := Car{}
	err := DB.DB.Where("ad_id = ?", adID).First(&found).Error
	return err == nil // if == nil then ad exists
}

func findUpdate(car *Car) {
	found := Car{}
	err := DB.DB.Where("ad_id = ?", car.AdID).First(&found).Error
	if err != nil {
		// Not found
		fmt.Println(err)
		// updateCar(car)
	}
}

func getPageCount(page io.Reader) int {
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		log.Fatal(err)
	}
	numStr := doc.Find(".result-count").Text()
	if len(numStr) < 2 {
		fmt.Println("no pages", numStr)
		return 0
	}
	numStr = numStr[1 : len(numStr)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		fmt.Println("cannot parse records", err)
		return 0
	}
	d := float64(num) / float64(20)
	return int(math.Ceil(d))
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func router() {
	// Create a new HTTP multiplexer
	mux := http.NewServeMux()

	// Add the pprof routes
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	// Start listening on port 8080
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(fmt.Sprintf("Error when starting or running http server: %v", err))
	}
}

func main() {
	go ExampleScrape()
	router()
}
