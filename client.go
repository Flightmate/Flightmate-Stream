package main

import (
	"crypto/tls"
	"encoding/binary"
	"flag"
	"github.com/Flightmate/Flightmate-Stream-Protobuf/click_packet"
	"github.com/Flightmate/Flightmate-Stream-Protobuf/search_packet"
	"github.com/go-restruct/restruct"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"hash/crc32"
	"log"
	"net"
	"os"
	"os/signal"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"io/ioutil"
)

var client_conn net.Conn
var client_connected = false
var packets_received = 0

// Parameters 
var print_json = false
var token string
var stdout = false

type Head struct {
	Checksum     uint32 //  0:4
	Body_Size    uint32 //  4:8
	Message_Type int8   //  9
}

func checkSum(checksum uint32, data []byte) bool {
	crc32q := crc32.MakeTable(0xD5828281)
	if checksum == crc32.Checksum(data, crc32q) {
		log.Println("checksum failed")
	} else {
		log.Println("checksum passed")
		return true
	}
	return false
}

func StartClient() {

	defer func() {
		if err := recover(); err != nil {
			log.Println("sendto: ", err)
			client_connected = false
			log.Println("retry connection")
			time.Sleep(5 * time.Second)
			StartClient()
		}
	}()

	config := &tls.Config{ServerName: "ai-stream.flightmate.com"}

	if !client_connected {
		// Connects to server
		client_conn, _ = tls.Dial("tcp", "ai-stream.flightmate.com:444", config) 
		client_connected = true
		log.Println("Connected to Poststation")
	}

	header_data := []byte{0, 0, 0, 0, 0, 0, 0, 0, 3}
	byte_token := []byte(token)
	concat_header_byte := append(header_data[:], byte_token[:]...)

	client_conn.Write(concat_header_byte)

	data := make([]byte, 4096)
	length, err := client_conn.Read(data) // Writes onto data
	if err != nil {
		log.Println(err.Error())
		client_connected = false
		log.Println("retry connection")
		time.Sleep(5 * time.Second)
	} else if string(data[:17]) == "Invalid OTA token" {
		log.Println("Invalid token")
	}

	header := Head{}
	unpack_err := restruct.Unpack(data[:9], binary.BigEndian, &header)
	if unpack_err != nil {
		log.Println(unpack_err.Error())
	}

	data = data[9:]

	if header.Message_Type == 1 {
		if checkSum(header.Checksum, data) {
			packets_received += 1
			searchPb := search_packet.Search_Packet{}

			err := proto.Unmarshal(data[:length], &searchPb)
			if err != nil {
				log.Println(err.Error())
			}

			// log.Printf("%+v\n", searchPb) // <-- Prints entire packet in Protobuf
			log.Printf("Received a search packet from %s to %s from %s", searchPb.From, searchPb.To, searchPb.Domain)

			if print_json || stdout {
				json_data := protobufToJSON(&searchPb)
				if stdout {
					fmt.Println(json_data)
				}  else if print_json {
					log.Println("json: ", json_data)
				}  	
			}
		}
	} else if header.Message_Type == 2 {
		if checkSum(header.Checksum, data) {
			packets_received += 1
			clickPb := click_packet.Click_Packet{}

			err := proto.Unmarshal(data[:length], &clickPb)
			if err != nil {
				log.Println(err.Error())
			}

			// log.Printf("%+v\n", clickPb) // <-- Prints entire packet in Protobuf
			log.Printf("Received a click packet from %s to %s from %s", clickPb.From, clickPb.To, clickPb.Domain)

			if print_json || stdout {
				json_data := protobufToJSON(&clickPb)
				if stdout {
					fmt.Println(json_data)
				}  else if print_json {
					log.Println("json: ", json_data)
				}  	
			}
		}
		log.Println("packets received: ", packets_received)
	}

	go func() {
		StartClient()
	}()

	// Prevents main (or startclient) from returning 
	runtime.Goexit() 
}

func notifyPoststationDisconnect() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-done:
			log.Printf("Got %s signal. Aborting...\n", sig)
			client_conn.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0, 5})
			os.Exit(1)
		}
	}()
}

func protobufToJSON(proto_message proto.Message) string {
	m := jsonpb.Marshaler{}
	json, _ := m.MarshalToString(proto_message)
	return json
}

func main() {
	// Prints line number when logging
	log.SetFlags(log.LstdFlags | log.Lshortfile) 

	token = "INSERT YOUR TOKEN HERE"

	// Makes it possible to use the token and print_json as cli parameters
	parameter_json := flag.Bool("print_json", false, "enable print json")
	parameter_token := flag.String("token", "", "insert your token")
	parameter_stdout := flag.Bool("stdout", false, "print json to stdout instead of log")

	flag.Parse()
	if *parameter_token != "" {
		token = *parameter_token
	} else if len(token) != 128 {
		log.Println(`You must insert your token for the client to be able to connect to the server. ` +
			`Do this by running "go run client.go -token YOUR_TOKEN_HERE" or by directly editing "token = 'INSERT YOUR TOKEN HERE'"`)
		os.Exit(3)
	}

	if *parameter_json {
		print_json = true
	}

	if *parameter_stdout {
		log.SetOutput(ioutil.Discard)
		stdout = true
	}

	notifyPoststationDisconnect()

	log.Println("Starting OTA tls client")
	StartClient()
}
