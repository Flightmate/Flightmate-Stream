package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/Flightmate/Flightmate-Stream-Protobuf/click_packet"
	"github.com/Flightmate/Flightmate-Stream-Protobuf/search_packet"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
	"os/signal"
	"syscall"
)

var client_conn net.Conn
var client_connected = false
var packets_received = 0

// Parameters
var token string
var print_json = false
var stdout = false

var parameter_host *string
var port int

var time_last_packet = time.Now()

type Header struct {
	Checksum     uint32 //  0:4
	Body_Size    uint32 //  4:8
	Message_Type int8   //  9
}

func checkSum(checksum uint32, data []byte) bool {
	crc32q := crc32.MakeTable(0xD5828281)
	return checksum == crc32.Checksum(data, crc32q)
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

func startClient() {	
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			client_connected = false
			log.Println("retry connection")
			time.Sleep(5 * time.Second)
			startClient()
		}
	}()

	config := &tls.Config{ServerName: *parameter_host}

	if !client_connected {
		var target = *parameter_host + ":" + strconv.Itoa(port)
		log.Printf("Connecting to Poststation %s", target)

		// Connects to server
		config := config
		client_conn, _ = tls.Dial("tcp", target, config)
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
		data := make([]byte, 9)                  // Add 9 because of length of header
		_, err := io.ReadFull(client_conn, data) // Writes onto data

		if err != nil {
			log.Println(err.Error())
			client_connected = false
			log.Println("retry connection")
			time.Sleep(5 * time.Second)
		}

		if string(data) == "isinvalid" {
			log.Println("Invalid token")
			os.Exit(3)
		}

		header := Header{}
		buf := bytes.NewBuffer(data)

		if err = binary.Read(buf, binary.BigEndian, &header); err != nil {
			log.Println(err.Error())
		}

		data = make([]byte, header.Body_Size) 
		_, err = io.ReadFull(client_conn, data) // Writes onto data

		if err != nil {
			log.Println(err.Error())
		}

		if checkSum(header.Checksum, data) {
			packets_received += 1

			time_last_packet = time.Now()

			if header.Message_Type == 1 {
				searchPb := search_packet.Search_Packet{}

				err := proto.Unmarshal(data, &searchPb)

				if err != nil {
					log.Println(err.Error())
				}

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
		startClient()
	}()

	// Prevents main (or startClient) from returning
	runtime.Goexit()
}

func protobufToJSON(proto_message proto.Message) string {
	json, err := protojson.Marshal(proto_message)

	if err != nil {
		log.Println(err.Error())
	}

	return string(json)
}

func parameterFunc() {
	// Makes it possible to use the token, print_json, and stdout as arguments
	// the syntax -flag x is allowed for non-boolean flags only, booleans must have =
	parameter_json := flag.Bool("print_json", false, "log responses as json")
	parameter_stdout := flag.Bool("stdout", false, "disable logging, only print json responses")

	parameter_token := flag.String("token", "", "insert your token 128 alphanumerical chars")

	parameter_host = flag.String("host", "ai-stream.flightmate.com", "target host you wish to connect to")
	parameter_port := flag.Int("port", 444, "host target port")

	flag.Parse()

	if *parameter_token != "" {
		token = *parameter_token
	} else if len(token) != 128 {
		flag.PrintDefaults()
		log.Fatal("Missing required token!")
	}

	if *parameter_json {
		print_json = true
	}

	if *parameter_stdout {
		log.SetOutput(ioutil.Discard)
		stdout = true
	}

	log.Println("port: ", *parameter_port)
	if *parameter_port != 0 {
		port = *parameter_port
	}
}

func main() {
	// Prints line number when logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	parameterFunc()

	// Graceful exit, close connection 
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-done:
			log.Println("Closing connection...: ", sig)
			os.Exit(0)
		}
	}()

	startClient()
}
