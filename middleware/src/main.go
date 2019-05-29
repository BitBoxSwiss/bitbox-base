package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	//"encoding/json"
	//"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	//"github.com/btcsuite/btcutil"
	"github.com/fiatjaf/lightningd-gjson-rpc"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	//"github.com/tidwall/gjson"
)

type middlewareInfoStruct struct {
	Blocks         int64   `json:"blocks"`
	Difficulty     float64 `json:"difficulty"`
	LightningAlias string  `json:"alias"`
}

func testBitcoinRPC(bitcoinRpcUser, bitcoinRpcPassword, bitcoinRpcPort string) middlewareInfoStruct {
	connCfg := rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         "127.0.0.1:" + bitcoinRpcPort,
		User:         bitcoinRpcUser,
		Pass:         bitcoinRpcPassword,
	}
	client, err := rpcclient.New(&connCfg, nil)
	if err != nil {
		log.Printf("error creating new btc client: %v", err)
	}
	//client is shutdown/deconstructed again as soon as this function returns
	defer client.Shutdown()

	//Get current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Printf("Unable to get Block count: %s", err.Error())
		blockCount = 0
	}

	blockChainInfo, err := client.GetBlockChainInfo()
	var difficulty = 0.0
	if err != nil {
		log.Printf("Unable to get blockchaininfo: %s", err)
	} else {
		difficulty = blockChainInfo.Difficulty
	}

	log.Printf("Block count: %d", blockCount)
	log.Printf("Blockchain info, blocks: %d, difficulty: %f", blockCount, difficulty)
	var returnStruct middlewareInfoStruct
	returnStruct.Blocks = blockCount
	returnStruct.Difficulty = difficulty
	return returnStruct
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan *middlewareInfoStruct)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func testCLightningRPC(lightningRpcPath string) string {
	ln := &lightning.Client{
		Path: lightningRpcPath,
	}

	nodeinfo, err := ln.Call("getinfo")
	var alias = "disconnected"
	if err != nil {
		log.Printf("getinfo error: %s" + err.Error())
	} else {
		alias = nodeinfo.Get("alias").String()
	}
	log.Print(alias)
	return alias
}

func rpcLoop(bitcoinRpcUser, bitcoinRpcPassword, bitcoinRpcPort, lightningRpcPath string) {
	for {
		var serviceInfo middlewareInfoStruct = testBitcoinRPC(bitcoinRpcUser, bitcoinRpcPassword, bitcoinRpcPort)
		serviceInfo.LightningAlias = testCLightningRPC(lightningRpcPath)
		go writer(&serviceInfo)
		time.Sleep(5 * time.Second)
	}
}

func main() {
	bitcoinRpcUser := flag.String("rpcuser", "rpcuser", "Bitcoin rpc user name")
	bitcoinRpcPassword := flag.String("rpcpassword", "rpcpassword", "Bitcoin rpc password")
	bitcoinRpcPort := flag.String("rpcport", "8332", "Bitcoin rpc port, localhost is assumed as an address")
	lightningRpcPath := flag.String("lightning-rpc-path", "/home/bitcoin/.lightning/lightning-rpc", "Path to the lightning rpc unix socket")
	flag.Parse()
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/ws", wsHandler)
	go rpcLoop(*bitcoinRpcUser, *bitcoinRpcPassword, *bitcoinRpcPort, *lightningRpcPath)
	go echo()

	log.Fatal(http.ListenAndServe(":8845", router))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK!!\n"))
	if err != nil {
		log.Print(err.Error())
	}
}

func writer(info *middlewareInfoStruct) {
	broadcast <- info
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// register client
	clients[ws] = true
}

func echo() {
	var i = 0
	for {
		i += 1
		val := <-broadcast
		blockinfo := fmt.Sprintf("%d %f %d %s", val.Blocks, val.Difficulty, i, val.LightningAlias)
		// send to every client that is currently connected
		fmt.Println(blockinfo)
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(blockinfo))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
