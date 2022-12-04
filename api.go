package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"crypto/tls"
)

const addr = "localhost:7070"
type City struct {
	ID       string
	Name     string `json:"name"`
	Location string `json:"location"`
}
func (c City) toJson() string {
	return fmt.Sprintf(`{"name":"%s","location":"%s"}`,
		c.Name,
		c.Location)
}


func main() {
	c := make(chan int)

	for i := 0; i < 5; i++ {
	
		go ciudades()
		go email()
		
	}

	for i := 0; i < 5; i++ {

		go estacionamiento2(i, c)
		desocupado := <-c
		fmt.Printf("el estacionamiento , numero %d, esta ocupado\n", desocupado)
	}
}

func estacionamiento2(id int, c chan int) {
	time.Sleep(3 * time.Second)
	fmt.Printf("el estacionamiento , numero %d, esta Desocupada\n", id)
	c <- id
}

func ciudades (){
	s := createServer(addr)
	go s.ListenAndServe()
	time.Sleep(time.Second * 5)

	cities, err := getCities()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Retrived cities: %v\n", cities)

	city, err := saveCity(City{"", "Paris", "France"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Saved city: %v\n", city)

	

}

func saveCity(city City) (City, error) {
	r, err := http.Post("http://"+addr+"/cities",
		"application/json",
		strings.NewReader(city.toJson()))
	if err != nil {
		return City{}, err
	}
	defer r.Body.Close()
	return decodeCity(r.Body)
}

func getCities() ([]City, error) {
	r, err := http.Get("http://" + addr + "/cities")
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return decodeCities(r.Body)
}

func decodeCity(r io.Reader) (City, error) {
	city := City{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&city)
	return city, err
}

func decodeCities(r io.Reader) ([]City, error) {
	cities := []City{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&cities)
	return cities, err
}

func createServer(addr string) http.Server {
	cities := []City{City{"1", "Prague", "Czechia"}, City{"2", "Bratislava", "Slovakia"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/cities", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		if r.Method == http.MethodGet {
			enc.Encode(cities)
		} else if r.Method == http.MethodPost {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			r.Body.Close()
			city := City{}
			json.Unmarshal(data, &city)
			city.ID = strconv.Itoa(len(cities) + 1)
			cities = append(cities, city)
			enc.Encode(city)
		}

	})
	return http.Server{
		Addr:    addr,
		Handler: mux,
	}

}

func email(){

	var email string
	fmt.Println("Enter username smtp: ")
	fmt.Scanln(&email)

	var pass string
	fmt.Println("Enter password smtp: ")
	fmt.Scanln(&pass)

	auth := smtp.PlainAuth("",
		email,
		pass,
		"smtp.gmail.com")

	c, err := smtp.Dial("smtp.gmail.com:587")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	config := &tls.Config{ServerName: "smtp.gmail.com"}

	if err = c.StartTLS(config); err != nil {
		panic(err)
	}

	if err = c.Auth(auth); err != nil {
		panic(err)
	}

	if err = c.Mail(email); err != nil {
		panic(err)
	}
	if err = c.Rcpt(email); err != nil {
		panic(err)
	}

	w, err := c.Data()
	if err != nil {
		panic(err)
	}

	msg := []byte("Este es el contenido del mensaje prro")
	if _, err := w.Write(msg); err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}
	err = c.Quit()

	if err != nil {
		panic(err)
	}
}