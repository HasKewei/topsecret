package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(localhost:3306)/logs")
	if err != nil {
		fmt.Println(err)
		return
	}
}

//build and lauch the server
func main() {
	server := http.Server{
		Addr:    ":5000",
		Handler: myHandler(),
	}

	server.ListenAndServe()
}

//build and return the server's handler
func myHandler() *mux.Router {
	mx := mux.NewRouter()
	mx.HandleFunc("/", Poster).Methods("POST")
	mx.HandleFunc("/{id}", GET).Methods("GET")
	return mx
}

//build the struct which holds the temp data

type Click struct {
	ID           string
	AdvertiserID int
	SiteID       int
	IP           string
	IosIfa       string
	GoogleAid    string
	WindowsAid   string
}

//the function that handles the GET method
//retrieve the json data form from database and the server to the browser
func GET(writer http.ResponseWriter, reader *http.Request) {
	///get the id from the name of maps
	id := mux.Vars(reader)["id"]

	//select data from sql databases according to the id
	row := db.QueryRow("SELECT id, advertiser_id, site_id, ip, ios_ifa FROM clicks WHERE id=?", id)

	//store the data from sql database to the temp struct
	var c Click
	err := row.Scan(&c.ID, &c.AdvertiserID, &c.SiteID, &c.IP, &c.IosIfa)
	if err == sql.ErrNoRows {
		io.WriteString(writer, "Error 404")
		writer.WriteHeader(http.StatusNotFound)
		return
	} else if errs != nil {
		io.WriteString(writer, "Error 500")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	//marshal the data from the temp struct to json
	bytes, errs := json.Marshal(&c)
	if errs != nil {
		fmt.Println(errs)
		return
	}

	//output the raw bytes to the browser
	io.WriteString(writer, string(bytes))
	writer.WriteHeader(http.StatusOK)

}

//the function which handle the post method
//post the json data from broswer to the server and sql databases
func Poster(w http.ResponseWriter, r *http.Request) {

	//get the raw bytes of input data
	bytes, errs := ioutil.ReadAll(r.Body)
	if errs != nil {
		fmt.Println(errs)
		return
	}

	//store the raw bytes to a temporary struct
	var point Click
	errs = json.Unmarshal(bytes, &point)
	if errs != nil {
		fmt.Println(errs)
		return
	}

	//generate a ramdom id for the post data and also get the ip address
	id := Id(point.AdvertiserID)
	ip := r.RemoteAddr

	//store the data from the struct to the sql databases
	_, errs = db.Exec("INSERT INTO clicks(id, advertiser_id, site_id, ip, ios_ifa, google_aid, windows_aid ) VALUES(?, ?, ?, ?, ?, ?, ?)",
		id, point.AdvertiserID, point.SiteID, ip, point.IosIfa, point.google_aid, point.windows_aid)

	if errs != nil {
		fmt.Println(errs)
		return
	} else {
		io.WriteString(w, id)
	}

}

//generate a random id to represent the unique id
func Id(adId int) string {
	t := time.Now()
	year, month, day := t.Date()
	var id = Hex(4) + "-" + strconv.Itoa(year) +
		strconv.Itoa(int(month)) + strconv.Itoa(day) + "-" + strconv.Itoa(adId)
	return id
}

//generate a random string encoded hex value with given byte
func Hex(chunks int) string {
	var buffer bytes.Buffer

	bytes := make([]byte, 4)
	for i := 0; i < chunks; i++ {
		binary.LittleEndian.PutUint32(bytes, mrand.Uint32())
		buffer.WriteString(hex.EncodeToString(bytes))
	}

	return buffer.String()
}
