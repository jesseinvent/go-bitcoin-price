package main

import (
	"encoding/json"
	"log"
	"time"
	"net/url"
	// "strconv"

	"github.com/gorilla/websocket"
)

type Socket struct {
	connection *websocket.Conn;
}

func (socket *Socket) connect() {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/btcusdt@aggTrade" }

	wsServerUrl := u.String();
	
	log.Printf("WS SERVER URL: %s", wsServerUrl);

	connection, _, err := websocket.DefaultDialer.Dial(wsServerUrl, nil);

	if err != nil {
		log.Fatal("Error connection to ws server:", err);
	}

	log.Print("Connection established to ws server...");

	connection.SetCloseHandler(func(code int, text string) error {

		log.Print("Server disconnected..");

		return nil;
	});

	connection.SetReadLimit(512);

	socket.connection = connection;
}

func (socket *Socket) readMessages(channel chan<- interface{}) {
	defer close(channel);

	for {
		var payload map[string]interface{};

		_, bytes, err := socket.connection.ReadMessage();

		if err != nil {
			log.Fatal("Error reading from stream:", err);
			break;
		}

		err = json.Unmarshal(bytes, &payload);

		if err != nil {
			log.Fatal(err);
			break;
		}

		price := payload["p"];
	
		channel <- price;
	}
}

var currentPrice interface{};

func main() {

	var socket Socket;
	
	socket.connect();

	defer socket.connection.Close();

	btcPriceChannel := make(chan interface{});

	done := make(chan bool, 1);

	go socket.readMessages(btcPriceChannel);

	go func() {
		defer close(done);
		for {
			currentPrice = <- btcPriceChannel;
		}
	}()

	for range time.Tick(time.Second * 3) {
		log.Printf("Current BTC Price: %v", currentPrice);
	}

	<- done;
}