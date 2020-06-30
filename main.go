package main

import (
	"fmt"
	"github.com/kaatinga/assets"
	"log"
	"net"
	"net/http"
)

const (
	port      = "8080"
	HTMLBegin = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>%encoder</title>\n</head>\n<body>"
	HTMLEnd   = "</body>\n</html>"
)

func index(w http.ResponseWriter, _ *http.Request) {

	title := "Приложение для домашней работы. Главная страница"

	_, err := fmt.Fprintf(w, HTMLBegin, title)
	if err != nil {
		log.Println(err)
	}

	_, err = fmt.Fprint(w, "Welcome! Add your name into URI '/hello' using GET method by parameter 'name'<br>Пример: <a href=/hello?name=Michael>/hello?name=Michael</a><br><a href=/shop>Магазин</a>")

	if err != nil {
		log.Println(err)
	}

	_, err = fmt.Fprint(w, HTMLEnd)
	if err != nil {
		log.Println(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := fmt.Fprint(w, "URI Get parameter 'name' is missing")
		if err != nil {
			log.Println(err)
		}

		return
	}

	_, err := fmt.Fprintf(w, "Welcome, %encoder!", name)
	if err != nil {
		log.Println(err)
	}
}

func (s *Shop) requestChecks(w http.ResponseWriter, r *http.Request) (basket *Basket) {

	basket = NewBasket()
	cookieName := "testShop"

	clearBasket := r.FormValue("clear")
	if clearBasket == "clear" {

		deleteCookie := &http.Cookie{
			Name:   cookieName,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		}

		http.SetCookie(w, deleteCookie)

		return
	}

	log.Println("action 'clear' is not found")

	var (
		goodID       = r.FormValue("goodid")
		goodIDUint16 uint16 // id товара
		ok           bool
	)

	if goodID != "" {
		log.Println("goodID is", goodIDUint16)
		goodIDUint16, ok = assets.StUint16(goodID)
		if !ok {
			log.Println("некорректный ID товара")
			return
		}
	} else {
		log.Println("goodID was not found")
	}

	var (
		goodAmount     = r.FormValue("goodamount")
		goodAmountByte byte
	)

	if goodID != "" {
		log.Println("goodAmount is", goodAmount)
		goodAmountByte, ok = assets.StByte(goodAmount)
		if !ok {
			log.Println("некорректное количество товара")
			return
		}
	} else {
		log.Println("goodAmount was not found")
	}

	var (
		basketCookie *http.Cookie
		err        error
	)

	// читаем куку
	basketCookie, err = r.Cookie(cookieName)
	if err != nil {
		log.Println("cookie", cookieName, " was not found")
	} else {

		err = s.encoder.Decode("cookie-name", basketCookie.Value, &basket.list)
		if err != nil {
			log.Println("Ошибка обработки данных куки:", err)
		}

		log.Println(basket.list)
	}

	var currentAmount byte

	currentAmount, ok = basket.list[goodIDUint16]
	if !ok {
		err = basket.AddGood(goodIDUint16, goodAmountByte)
		if err != nil {
			log.Println(err)
		}
	} else {
		err = basket.AddGood(goodIDUint16, currentAmount + goodAmountByte)
		if err != nil {
			log.Println(err)
		}
	}

	if encodedBasket, err := s.encoder.Encode("cookie-name", basket.list); err == nil {

		formCookie := &http.Cookie{
			Name:     cookieName,
			Value:    encodedBasket,
			Path:     "/shop",
			MaxAge:   3000,                    // 50 минут
			Secure:   false,                   // yet 'false' as TLS is not used
			HttpOnly: true,                    // 'true' secures from XSS attacks
			SameSite: http.SameSiteStrictMode, // base CSRF attack protection
		}

		http.SetCookie(w, formCookie)
	} else {
		log.Println("Ошибка маршализации:", err)
	}

	return
}

func (s *Shop) shop(w http.ResponseWriter, r *http.Request) {

	basket := NewBasket()

	if r.Method == "POST" {
		basket = s.requestChecks(w, r)
		log.Println(basket)
	} else {
		log.Println("No POST data available")
	}

	var (
		title    = ShopName
		goodList string
	)

	_, err := fmt.Fprintf(w, HTMLBegin, title)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprint(w, "<h2>Список товаров</h2>")

	// формируем часть кода со списком товаров
	goodList, err = s.GetGoods()
	if err != nil {
		log.Println(err)
	}

	_, err = fmt.Fprint(w, goodList)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprint(w, "<h2>Ваша карзина</h2>")

	var total, cost uint64
	for key, value := range basket.list {

		good, ok := s.GetGood(key)
		if ok {
			cost = good.price*uint64(value)
			_, err = fmt.Fprint(w, "Товар: ", good.name, ", кол-во: ", value,", цена: ", cost , "<br>")
			if err != nil {
				log.Println(err)
			}

		total = total + cost

		} else {
			log.Println("Товар не найден")
		}
	}

	// Итого
	_, err = fmt.Fprint(w, "<p>ИТОГО:" , total, "</p>")
	if err != nil {
		log.Println(err)
	}


	// очистка куки
	_, err = fmt.Fprint(w, "<form action=/shop method=post><input type=hidden name=clear value=clear><button type=submit>Очистить карзину</button></form>")
	if err != nil {
		log.Println(err)
	}

	// формирование заказа
	_, err = fmt.Fprint(w, "<form action=/order method=get><button type=submit>Сформировать заказ</button></form>")
	if err != nil {
		log.Println(err)
	}

	fmt.Fprint(w, HTMLEnd)

}

func (s *Shop) order(w http.ResponseWriter, r *http.Request) {
	title := "A handler to make order"

	orderID := r.URL.Query().Get("order")

	if orderID == "" {
		log.Println("Ошибка при ввода номера заказа")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, ok := assets.StUint16(orderID)
	if !ok {
		log.Println("Ошибка при ввода номера заказа. Допустимы только числа")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := fmt.Fprintf(w, HTMLBegin, title)
	if err != nil {
		log.Println(err)
	}

	_, err = fmt.Fprintf(w, "Order ID is %encoder", orderID)
	if err != nil {
		log.Println(err)
	}

	//err = encoder.SendEmail(fmt.Sprintf("Спасибо за то что Вы сделали заказ. Ваш номер заказа: %encoder", orderID))
	//if err != nil {
	//	log.Println(err)
	//}

	_, err = fmt.Fprint(w, HTMLEnd)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	var (
		err error
	)

	// Инициализация товаров
	theShop := NewShop()
	err = theShop.AddGood("Яблоко", "Кг.", 100)
	if err != nil {
		log.Println(err)
	}

	err = theShop.AddGood("Груша", "Кг.", 200)
	if err != nil {
		log.Println(err)
	}

	err = theShop.AddGood("Шоколад", "Шт.", 300)
	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/shop", theShop.shop)
	http.HandleFunc("/order", theShop.order)

	fmt.Println("Server is running...")
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}
