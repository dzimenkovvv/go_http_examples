package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type Payment struct {
	Description string `json:"description"`
	USD         int    `json:"usd"`
	FullName    string `json:"fullName"`
	Addres      string `json:"addres"`
}

func (p Payment) Print() {
	fmt.Println("Description:", p.Description)
	fmt.Println("USD:", p.USD)
	fmt.Println("Full Name:", p.FullName)
	fmt.Println("Addres:", p.Addres)
}

var money int = 1000
var mtx = sync.Mutex{}
var paymentHistory = make([]Payment, 0)

func payHandler(w http.ResponseWriter, r *http.Request) {
	var payment Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payment.Print()

	mtx.Lock()
	if money-payment.USD >= 0 {
		money -= payment.USD

		w.Write([]byte("Покупка успешно совершена! \nСумма покупки: " + strconv.Itoa(payment.USD) + " USD!\nОстаток баланса: " + strconv.Itoa(money) + " USD!"))

		paymentHistory = append(paymentHistory, payment)
		fmt.Println(paymentHistory)
	} else {
		w.WriteHeader(http.StatusBadRequest)

		w.Write([]byte("Недостаточно средств!"))
		fmt.Println("Недостаточно средств!")
	}
	mtx.Unlock()
}

func main() {
	http.HandleFunc("/pay", payHandler)

	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Server HTTP error")
		return
	}
}
