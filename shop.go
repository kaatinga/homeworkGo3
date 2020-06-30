package main

import (
	"errors"
	"github.com/gorilla/securecookie"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

const ShopName string = "Тестовый магазин"

// объекты и методы для хранения паролей
type Shop struct {
	mx      sync.RWMutex
	goods   Goods
	orders  Orders
	encoder *securecookie.SecureCookie
}

func NewShop() *Shop {

	return &Shop{
		goods: Goods{
			list:    make(map[uint16]Good, 100),
			entryID: 0,
		},
		orders:  make(map[uint32]Order, 10),
		encoder: securecookie.New([]byte("testshoptestshop"), []byte("shoptestshoptest")),
	}
}

func (s *Shop) AddGood(name, unit string, price uint64) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if name == "" || unit == "" {
		return errors.New("incorrect input data")
	}

	if price == 0 {
		return errors.New("incorrect price")
	}

	s.goods.list[s.goods.entryID] = Good{
		name:  name,
		price: price,
		unit:  unit,
	}

	s.goods.entryID = s.goods.entryID + 1

	return nil
}

func (s *Shop) DeleteGood(goodID uint16) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, ok := s.goods.list[goodID]
	if ok {
		delete(s.goods.list, goodID)
		s.goods.entryID = goodID
		return nil
	}

	return errors.New("there is no such a good in the db")
}

func (s *Shop) GetGoods() (goods string, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for key, value := range s.goods.list {
		goods = strings.Join([]string{goods, "<p>", value.name, ", цена за ", value.unit, " ", strconv.FormatUint(value.price, 10), " руб., <form action=/shop method=post><input type=hidden name=goodamount value=1><input type=hidden name=goodid value=", strconv.Itoa(int(key)), "><button type=submit>Добавить</button></form></p>"}, "")
	}
	return
}

func (s *Shop) GetGood(goodID uint16) (good Good, ok bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	good, ok = s.goods.list[goodID]
	if ok {
		return
	}

	return good, false
}

type Good struct {
	id    uint16
	name  string
	price uint64
	unit  string
}

type Goods struct {
	list    map[uint16]Good
	entryID uint16
}

type Client struct {
	ID    uint16
	Name  string
	Email string
}

type Order struct {
	OrderID uint32
	*Client
	*Basket
}

// количество = value, orderID = key
type Basket struct {
	mx   sync.RWMutex
	list map[uint16]uint8
}

func NewBasket() *Basket {
	return &Basket{
		list: make(map[uint16]byte, 10),
	}
}

func (b *Basket) AddGood(goodID uint16, amount uint8) error {
	b.mx.Lock()
	defer b.mx.Unlock()

	b.list[goodID] = amount

	return nil
}

func GenOrderID() uint32 {
	return rand.Uint32()
}

func (s *Shop) NewOrder(basket *Basket, client *Client) Order {
	s.mx.RLock()
	defer s.mx.RUnlock()

	var (
		ok      bool
		orderID uint32
	)

	for {
		orderID = GenOrderID()
		_, ok = s.orders[orderID]
		if !ok {
			break
		}
	}

	return Order{OrderID: orderID, Basket: basket, Client: client}
}

type Orders map[uint32]Order
