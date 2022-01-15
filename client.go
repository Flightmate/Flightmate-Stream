package main

import (
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/Flightmate/Flightmate-Stream-Protobuf/click_packet"
	"github.com/Flightmate/Flightmate-Stream-Protobuf/search_packet"
	"github.com/go-restruct/restruct"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"time"
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
		return true
	}
	return false
}

func printLogic(packet proto.Message) {
	if print_json || stdout {
		json_data := protobufToJSON(packet)

		if stdout {
			fmt.Println(json_data)
		} else if print_json {
			log.Println("json: ", json_data)
		}
	}
}

func StartClient() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			client_connected = false
			log.Println("retry connection")
			time.Sleep(5 * time.Second)
			StartClient()
		}
	}()

	
	if !client_connected {
		// Connects to server
		config := &tls.Config{ServerName: "ai-stream.flightmate.com"}
		client_conn, _ = tls.Dial("tcp", "ai-stream.flightmate.com:444", config)
		client_connected = true
		log.Println("Connected to Poststation")
		header_data := []byte{0, 0, 0, 0, 0, 0, 0, 0, 3}
		byte_token := []byte(token)
		concat_header_byte := append(header_data[:], byte_token[:]...)

		_, err := client_conn.Write(concat_header_byte)

		if err != nil {
			log.Println(err.Error())
		}
	}

	if client_connected {
		size_of_bytes := make([]byte, 8)
		_, err := io.ReadFull(client_conn, size_of_bytes)

		if err != nil {
			log.Println(err.Error())
		} 

		size_of_proto := int64(binary.LittleEndian.Uint64(size_of_bytes[:8]))

		// if size_of_proto == 100000 {
		// 	log.Println("Invalid token")
		// } 

		data := make([]byte, size_of_proto+9)   // Add 9 because of length of header
		_, err = io.ReadFull(client_conn, data) // Writes onto data

		if err != nil {
			log.Println(err.Error())
			client_connected = false
			log.Println("retry connection")
			time.Sleep(5 * time.Second)
		} else {
			// log.Println(data)
			// log.Println(string(data))
		}/*if string(data[:7]) == "Invalid" {
			log.Println("Invalid token")
		}*/

		header := Head{}
		unpack_err := restruct.Unpack(data[:9], binary.BigEndian, &header)

		if unpack_err != nil {
			log.Println(unpack_err.Error())
		}

		data = data[9:]

		if checkSum(header.Checksum, data) {
			packets_received += 1
			if header.Message_Type == 1 {
				searchPb := search_packet.Search_Packet{}

				err := proto.Unmarshal(data, &searchPb)
				if err != nil  {
					log.Println(err.Error())
				}

				// log.Printf("%+v\n", searchPb) // <-- Prints entire packet in Protobuf
				log.Printf("Received a search packet from %s to %s from %s", searchPb.From, searchPb.To, searchPb.Domain)

				printLogic(&searchPb)
			} else if header.Message_Type == 2 {
				clickPb := click_packet.Click_Packet{}

				err := proto.Unmarshal(data, &clickPb)
				if err != nil {
					log.Println(err.Error())
				}

				// log.Printf("%+v\n", clickPb) // <-- Prints entire packet in Protobuf
				log.Printf("Received a click packet from %s to %s from %s", clickPb.From, clickPb.To, clickPb.Domain)

				printLogic(&clickPb)
			}
		}
		// log.Println("packets received: ", packets_received)
	}

	go func() {
		StartClient()
	}()

	// Prevents main (or startclient) from returning
	runtime.Goexit()
}

func protobufToJSON(proto_message proto.Message) string {
	m := jsonpb.Marshaler{}
	json, _ := m.MarshalToString(proto_message)
	return json
}

func parameterFunc() {
	// Makes it possible to use the token, print_json, and stdout as arguments
	parameter_json := flag.Bool("print_json", false, "enable print json")
	parameter_token := flag.String("token", "", "insert your token")
	parameter_stdout := flag.Bool("stdout", false, "print json to stdout instead of log")

	flag.Parse()

	if *parameter_token != "" {
		token = *parameter_token
	} else if len(token) != 128 {
		log.Println(`You must insert your token for the client to be able to connect to the server. ` +
			`Specify with the argument --token YOUR_TOKEN_HERE" (or by directly editing "token = 'INSERT YOUR TOKEN HERE' if you have cloned the repo)"`)
		os.Exit(3)
	}

	if *parameter_json {
		print_json = true
	}

	if *parameter_stdout {
		log.SetOutput(ioutil.Discard)
		stdout = true
	}
}

func main() {
	// Prints line number when logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	token = "INSERT YOUR TOKEN HERE"

	parameterFunc()

	log.Println("Starting OTA tls client")
	StartClient()
}
