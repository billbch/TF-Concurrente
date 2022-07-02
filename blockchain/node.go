package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type Tweet struct {
	User_id string `json:"user_id"`
	Message string `json:"message"`
	Publish_date string `json:"publish_date"`
}

type Block struct {
	Pos       int             `json:"pos"`
	Data      Tweet						`json:"data"`
	Timestamp string          `json:"timestamp"`
	Hash      string          `json:"hash"`
	PrevHash  string          `json:"prevhash"`
}

type Blockchain struct {
	Blocks []Block 		`json:"blocks"`
	NewerBlock Block 	`json:"newerblock"`
}

var localBlockchain Blockchain


type Frame struct {
	Cmd    string   `json:"cmd"`
	Sender string   `json:"sender"`
	Data   []string `json:"data"`
}

type Info struct {
	nextNode string
	nextNum  int
	imFirst  bool
	cont     int
}

type InfoCons struct {
	cont map[string][]string
	sum int
}

var (
	host         string
	myNum        int
	chRemotes    chan []string
	chInfo       chan Info
	chCons       chan InfoCons
	readyToStart chan bool
	participants int
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) == 1 {
		log.Println("Hostname not given")
	} else {
		host = os.Args[1]
		chRemotes = make(chan []string, 1)
		chInfo = make(chan Info, 1)
		chCons = make(chan InfoCons, 1)
		readyToStart = make(chan bool, 1)

		chRemotes <- []string{}
		if len(os.Args) >= 3 {
			connectToNode(os.Args[2])
		}
		if len(os.Args) == 4 {
			switch os.Args[3] {
			case "agrawalla":
				go startAgrawalla()
			case "consensus":
				go startConsensus()
			case "new_block":
				datos := Tweet {
					User_id: "tilin",
					Message: "volvio Abimael",
					Publish_date: "30-06-2022",
				}
				go startNewBlock(datos)
			}
		}
		server()
	}
}
func CreateBlock(datos Tweet) Block {
	prevBlock := localBlockchain.NewerBlock

	// si el pre bloque no existe.
	if (prevBlock == Block{}) {
		prevBlock = Block {
			Pos: -1,
			Data: Tweet{},
			Timestamp: "00:00:00",
			Hash: "0000",
			PrevHash: "",
		}
	}

	block := Block{}
	block.Pos = prevBlock.Pos + 1
	// se usa un tiempo fijo
	//block.Timestamp = time.Now().String()
	block.Timestamp = "00:00:00"
	block.Data = datos
	block.PrevHash = prevBlock.Hash
	block.Hash = block.generateHash()

	return block
}
func (block Block) generateHash() string {
	// obtener valoe de la Data
	bytes, _ := json.Marshal(block.Data)
	// concatenar el conjunto de datos
	data := strconv.Itoa(block.Pos) + block.Timestamp + string(bytes) + block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}
func startNewBlock(datos Tweet) {
	remotes := <-chRemotes
	chRemotes <- remotes

	out, _ := json.Marshal(datos)
	data := string(out)

	frame := Frame{"add_block", host, []string{data}}
	for _, remote := range remotes {

		send(remote, frame, nil)
	}
	handleAddBlock(&frame)

	go startConsensus()
}
func startAgrawalla() {
	remotes := <-chRemotes
	chRemotes <- remotes
	for _, remote := range remotes {
		send(remote, Frame{"agrawalla", host, []string{}}, nil)
	}
	handleAgrawalla()
}
func startConsensus() {
	remotes := <-chRemotes
	for _, remote := range remotes {
		log.Printf("%s: notifying %s\n", host, remote)
		send(remote, Frame{"consensus", host, []string{}}, nil)
	}
	chRemotes <- remotes
	handleConsensus()
}
func connectToNode(remote string) {
	remotes := <-chRemotes
	remotes = append(remotes, remote)
	chRemotes <- remotes
	if !send(remote, Frame{"hello", host, []string{}}, func(cn net.Conn) {
		dec := json.NewDecoder(cn)
		var frame Frame
		dec.Decode(&frame)
		remotes := <-chRemotes
		remotes = append(remotes, frame.Data...)
		chRemotes <- remotes
		log.Printf("%s: friends: %s\n", host, remotes)

		// peticion blockchain
		var received_blockchain Blockchain

		dec.Decode(&frame)
		data := frame.Data[0]
		ba := []byte(data)

		json.Unmarshal(ba, &received_blockchain)

		localBlockchain = received_blockchain
		
		// imprimir blockchain
		out, _ := json.Marshal(localBlockchain.Blocks)
		log.Printf("%s: Estoy en ConnectToNode - Mi blockchain es: %s.\n", host, string(out))

	}) {
		log.Printf("%s: unable to connect to %s\n", host, remote)
	}
}
func send(remote string, frame Frame, callback func(net.Conn)) bool {
	if cn_exterior, err := net.Dial("tcp", remote); err == nil {
		defer cn_exterior.Close()
		enc := json.NewEncoder(cn_exterior)
		enc.Encode(frame)
		if callback != nil {
			callback(cn_exterior)
		}
		return true
	} else {
		log.Printf("%s: can't connect to %s\n", host, remote)
		idx := -1
		remotes := <-chRemotes
		for i, rem := range remotes {
			if remote == rem {
				idx = i
				break
			}
		}
		if idx >= 0 {
			remotes[idx] = remotes[len(remotes)-1]
			remotes = remotes[:len(remotes)-1]
		}
		chRemotes <- remotes
		return false
	}
}
func server() {
	if ln, err := net.Listen("tcp", host); err == nil {
		defer ln.Close()
		log.Printf("Listening on %s\n", host)
		for {
			if cn, err := ln.Accept(); err == nil {
				go fauxDispatcher(cn)
			} else {
				log.Printf("%s: Can't accept connection.\n", host)
			}
		}
	} else {
		log.Printf("Can't listen on %s\n", host)
	}
}
func fauxDispatcher(cn net.Conn) {

	defer cn.Close()

	dec := json.NewDecoder(cn)
	frame := &Frame{}
	dec.Decode(frame)
	
	log.Printf("%s: fauxDispatcher :: %s - %s - %s\n", host, frame.Cmd, frame.Sender, frame.Data)

	switch frame.Cmd {
		case "hello":
			handleHello(cn, frame)
		case "add":
			handleAdd(frame)
		case "agrawalla":
			handleAgrawalla()
		case "num":
			handleNum(frame)
		case "start":
			handleStart()
		case "consensus":
			handleConsensus()
		case "vote":
			handleVote(frame)
		case "add_block":
			handleAddBlock(frame)
		case "new_block":
			var datos Tweet
			data := frame.Data[0]

			ba := []byte(data)

			json.Unmarshal(ba, &datos)
			
			go startNewBlock(datos)
		case "get_blockchain":
			handleGetBlockchain(cn, frame)
		case "api_connect":
			ApiDispatcher(frame, cn)
		default:
			log.Printf("%s: entro en faux - en default\n", host)
	}
}
func ApiDispatcher(frame *Frame, cn net.Conn) {

	ApiFrame := frame

	for {
		switch ApiFrame.Cmd {
			case "api_new_block":
				var datos Tweet
				data := ApiFrame.Data[0]

				ba := []byte(data)

				json.Unmarshal(ba, &datos)
				
				go startNewBlock(datos)
			case "api_get_blockchain":
				handleApiGetBlockchain(cn)

				//fauxDispatcher(cn)
			case "api_connect":
				log.Printf("%s: Conexion con API iniciada.\n", host)
		}

		dec := json.NewDecoder(cn)
		
		ApiFrame = &Frame{}
		
		dec.Decode(ApiFrame)
				
		log.Printf("%s: ApiDispatcher :: %s - %s - %s\n", host, ApiFrame.Cmd, ApiFrame.Sender, ApiFrame.Data)
	}
}
func handleApiGetBlockchain(cn net.Conn) {
	out, _ := json.Marshal(localBlockchain.Blocks)
	io.WriteString(cn,string(out))
}
func handleAddBlock(frame *Frame) {
	var (
		datos Tweet
		newBlock Block
	)
	data := frame.Data[0]
	ba := []byte(data)

	json.Unmarshal(ba, &datos)

	newBlock = CreateBlock(datos)
	
	// agregamos el bloque a la cadena local
	localBlockchain.Blocks = append(localBlockchain.Blocks, newBlock)
	localBlockchain.NewerBlock = newBlock

	// imprimir blockchain
	out, _ := json.Marshal(localBlockchain.Blocks)
	log.Printf("%s: Estoy en AddBlock - Mi blockchain es: %s.\n", host, string(out))
}
func handleGetBlockchain(cn net.Conn, frame *Frame) {
	enc := json.NewEncoder(cn)

	// enviamos nuestro blockchain
	out, _ := json.Marshal(localBlockchain)
	data := string(out)
	enc.Encode(Frame{"<response>", host, []string{data}})
}
func handleHello(cn net.Conn, frame *Frame) {
	enc := json.NewEncoder(cn)
	remotes := <-chRemotes
	enc.Encode(Frame{"<response>", host, remotes})

	// enviamos nuestro blockchain
	out, _ := json.Marshal(localBlockchain)
	data := string(out)
	enc.Encode(Frame{"<response>", host, []string{data}})


	notification := Frame{"add", host, []string{frame.Sender}}
	for _, remote := range remotes {
		send(remote, notification, nil)
	}
	remotes = append(remotes, frame.Sender)
	log.Printf("%s: friends: %s\n", host, remotes)
	chRemotes <- remotes
}
func handleAdd(frame *Frame) {
	remotes := <-chRemotes
	remotes = append(remotes, frame.Data...)
	log.Printf("%s: friends: %s\n", host, remotes)
	chRemotes <- remotes
}
func handleAgrawalla() {
	myNum = rand.Intn(1000000000)
	log.Printf("%s: my number is %d\n", host, myNum)
	msg := Frame{"num", host, []string{strconv.Itoa(myNum)}}
	remotes := <-chRemotes
	chRemotes <- remotes
	for _, remote := range remotes {
		send(remote, msg, nil)
	}
	chInfo <- Info{"", 1000000001, true, 0}
}
func handleNum(frame *Frame) {
	if num, err := strconv.Atoi(frame.Data[0]); err == nil {
		info := <-chInfo
		//log.Printf("from %v\n", frame)
		if num > myNum {
			if num < info.nextNum {
				info.nextNum = num
				info.nextNode = frame.Sender
			}
		} else {
			info.imFirst = false
		}
		info.cont++
		chInfo <- info
		remotes := <-chRemotes
		chRemotes <- remotes
		if info.cont == len(remotes) {
			if info.imFirst {
				log.Printf("%s: I'm first!\n", host)
				criticalSection()
			} else {
				readyToStart <- true
			}
		}
	} else {
		log.Printf("%s: can't convert %v\n", host, frame)
	}
}
func handleStart() {
	<-readyToStart
	criticalSection()
}
func handleConsensus() {
	/*fmt.Print("A o B, elige: ")
	fmt.Scanf("%s\n", &op)*/
	InfoCons := InfoCons{}
	InfoCons.cont = make(map[string][]string)//map[string][]string{}
	InfoCons.sum = 0
	
	// current block
	block := localBlockchain.NewerBlock

	InfoCons.cont[block.Hash] = append(InfoCons.cont[block.Hash], host)
	InfoCons.sum++

	chCons <- InfoCons

	remotes := <-chRemotes
	participants = len(remotes) + 1

	for _, remote := range remotes {
		log.Printf("%s: sending %s to %s\n", host, block.Hash, remote)
		send(remote, Frame{"vote", host, []string{block.Hash}}, nil)
	}
	chRemotes <- remotes
}
func handleVote(frame *Frame) {
	
	vote := frame.Data[0]
	InfoCons := <-chCons

	/*if _, ok := InfoCons.cont[vote]; ok {
    InfoCons.cont[vote]++
	} else {
		InfoCons.cont[vote] = 1
	}*/
	InfoCons.cont[vote] = append(InfoCons.cont[vote], frame.Sender)
	InfoCons.sum++

	chCons <- InfoCons
	log.Printf("%s: %s voted %s\n", host, frame.Sender, vote)
	if InfoCons.sum == participants {
		handleResult()
	}
}
func handleResult() {
	var (
		maxHash string
		maxCont int
		medium int
	)

	InfoCons := <-chCons

	for key, value := range InfoCons.cont {
		length := len(value)
		if (length > maxCont) {
			maxHash = key
			maxCont = length
		}
	}
	medium = (participants / 2) + 1

	if (maxCont >= medium) {
		log.Printf("%s: el hash mas votado es %s con %d votos de %d participantes.\n", host, maxHash, maxCont, participants)
	} else {
		log.Printf("%s: Hubo un empate en las votaciones.\n", host)
	}
	
	// current block
	block := localBlockchain.NewerBlock

	// imprimir blockchain
	out, _ := json.Marshal(localBlockchain.Blocks)
	log.Printf("%s: Mi blockchain es: %s.\n", host, string(out))

	// el localBlockchain es correcto?
	if (maxHash == block.Hash) {
		log.Printf("%s: Mi blockchain es correcto.\n", host)
	} else {
		log.Printf("%s: Mi blockchain es incorrecto. Solicito copia de blockchain correcto.\n", host)
		
		if len(InfoCons.cont[maxHash]) > 0 {
			
			//remote := InfoCons.cont[maxHash][0]
			// recorremos los remotes que poseen hash mas votado.
			for _, remote := range InfoCons.cont[maxHash] {
				if send(remote, Frame{"get_blockchain", host, []string{}}, func(cn net.Conn) {
				dec := json.NewDecoder(cn)
				var frame Frame
				// peticion blockchain
				var received_blockchain Blockchain

				dec.Decode(&frame)
				data := frame.Data[0]
				ba := []byte(data)
				json.Unmarshal(ba, &received_blockchain)

				localBlockchain = received_blockchain
				
				// imprimir blockchain
				out, _ := json.Marshal(localBlockchain.Blocks)
				log.Printf("%s: Estoy en handleResult - Mi blockchain es: %s.\n", host, string(out))

			}) {
				break
			} else {
				log.Printf("%s: unable to connect to %s\n", host, remote)
			}
			}
		}
	}
}
func criticalSection() {
	log.Printf("%s: my time has come!\n", host)
	info := <-chInfo
	if info.nextNode != "" {
		log.Printf("%s: letting %s start\n", host, info.nextNode)
		send(info.nextNode, Frame{"start", host, []string{}}, nil)
	} else {
		log.Printf("%s: I was the last one :(\n", host)
	}
}
