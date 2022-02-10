# **Flightmate AI-Stream Documentation**

This version of the AI-stream is rewritten from Python to Golang to be able to handle heavier loads. 

The server now uses [Protobuf](https://www.wikiwand.com/en/Protocol_Buffers) for faster communication, but the client can convert it into JSON (e.g. by using the optional parameter `--print_json=true`). You can also print exclusively JSON to stdout by using the flag `--stdout=true`. We have also upgraded to using TLS instead of SSL. 

If you want to run by cloning the repo (`git clone -b Flightmate-Stream-2022 --single-branch https://github.com/Flightmate/Flightmate-Stream/`):
- Install Golang [here](https://go.dev/doc/install) 
- Run with `go run client.go --token=YOUR_TOKEN_HERE` 
- You can also fwe edit `token = "INSERT YOUR TOKEN HERE` directly in client.go 

If you use the downloaded binaries (found under [Releases](/releases/latest)). (Note that you might get a "possible virus" warning (this will be fixed in later versions)): 
- Navigate to the file's location
- You might have to run chmod +x [filename] to change permissions 
- Run `./filename --token=YOUR_TOKEN_HERE` 

You can build the files yourself using: 
- env GOOS=windows  GOARCH=386 go build -o executables/streamclient-windows-386.exe client.go
- env GOOS=windows  GOARCH=amd64 go build -o executables/streamclient-windows.exe client.go
- env GOOS=darwin GOARCH=386 go build -o executables/streamclient-macOS-386 client.go
- env GOOS=darwin GOARCH=amd64 go build -o executables/streamclient-macOS client.go
- env GOOS=linux GOARCH=386 go build -o executables/streamclient-linux-386 client.go
- env GOOS=linux GOARCH=amd64 go build -o executables/streamclient-linux client.go

## **System description**

Flightmate AI-Stream is a stream of data containing all results displayed to the users in the result list, at flygresor.se. Results are replicated in real-time by the server and available to Flygresor.se clients. The stream can be tapped into and the data can be stored in the customer's AI or BI systems to analyse the results and use it to improve content or pricing in their API responses to flygresor.se. The customer's results will be identified with a customer name code and the other customer's names will be anonymised in the stream.

## **Server information**

<table>
  <tr>
    <td>Hostname</td>
    <td>ai-stream.flightmate.com</td>
  </tr>
  <tr>
    <td>Port</td>
    <td>444</td>
  </tr>
  <tr>
    <td>Protocol</td>
    <td>TCP</td>
  </tr>
  <tr>
    <td>Encryption</td>
    <td>TLS</td>
  </tr>
</table>


## **Visualization of client/server communication from start to stream**

![image alt text](https://lh4.googleusercontent.com/bCJEILOU0trwFdwSvAqn_V4hN89RLQ1CE98mjhAiC5ioDwMFV79fBAh46tpI64qBztEikxuiYXilRH3B-NSY1Q5udEheopR99_PVdRZ1jDNo9nPCF-iBM-ojFscPajCxFpKfjwkO)

## **Packet information shorthand**

<table>
  <tr>
    <td>Byte order</td>
    <td>Big endian</td>
  </tr>
  <tr>
    <td>Encoding</td>
    <td>UTF-8</td>
  </tr>
  <tr>
    <td>Header size</td>
    <td>9 bytes</td>
  </tr>
  <tr>
    <td>Header types</td>
    <td>[Unsigned long (4B), unsigned long (4B), unsigned char(1B)]</td>
  </tr>
  <tr>
    <td>Header content</td>
    <td>[Checksum, body size, packet type]</td>
  </tr>
    <tr>
    <td>Body size</td>
    <td>Byte length of Protobuf</td>
  </tr>
  <tr>
    <td>Packet type 1</td>
    <td>JSON message with search data</td>
  </tr>
  <tr>
    <td>Packet type 2</td>
    <td>JSON message with click out data.</td>
  </tr>
  <tr>
    <td>Packet type 3</td>
    <td>Authentication package containing OTAs 128B long token.</td>
  </tr>
</table>


## **Authentication packets**

To access the data stream an authentication is required. This will allow you access to the datastream and unmask the OTA signatures your allowed to see. If an authentication packet is malformed or does not match the server will drop the client. The authentication packet is a character sequence of 128 bytes and is supplied by Flightmate to the OTA. Please contact Valdemar at valle@flygresor.se for this.

## **"Continuous stream of data packets" breakdown**

The first 9 bytes of each package contains the package header. This header contains information the length of the package body, the package type and a checksum to verify the package’s integrity. It’s constructed like this: 

[(int) checksum (4 bytes), (int) body size (4 bytes), (int) packet type (1 byte)]

The checksum is made from the whole packet so to verify it you need to first unpack the header then pack it down without the checksum like this:

[(int) body size (4 bytes), (int) packet type (1 byte)]

Then you can verify the package with the checksum using the crc32 algorithm.

There are two types of packet you can receive from the stream. Packet type 1 contains search results showed to the user, and packet type 2 contains information about when a user clicks out on a trip and gets transferred to a OTA. 

### Click packet:

The click packet is sent out each time a user clicks out from one of the sites and it contains the following data:

**price:** The price of the flight the user clicked on.

**name:** The name of the OTA providing the result the user clicked on. This is masked if you don’t have access to that OTA.

**searchIdentifier:** This is a hash of some data specific to the trip the user clicked on making it possible to match it to a specific search result.

**domain:** The domain name of the site the outclick came from.

**isBaggageIncluded:** A boolean indicating whether the user has activated the baggage included filter or not.

**legs:** A json list of all searched trip legs. Each leg is a json object that include the following fields:

* **From**: An iata code. Example "ARN".

* **To**: An iata code.

* **Date**: The leave date specified by the user. Example 2018-12-30

**leaveDate:** The date of the flight leaving the departure location.

**homeDate:** The date of departure for the return flight.

**adults:** The number of adults specified in the search.

**childrenAges:** The ages of the children specified by the user when the search is made.

**youthAges:** The ages of the youths (12-25) specified by the user when the search is made.

**device:** The type of the user device. Could be one of these three values: "DESKTOP", “TABLET” or “MOBILE”.

### Search packet:

The search packets are sent each time a user display a search result(this includes users opening results from top-list and last minute) and contains the following data:

**to:** The IATA code of the airport the flight is arriving to.

**from:** The IATA code of the airport the flight is departing from.

**leaveDate:** The date of the flight leaving the departure location.

**homeDate:** The date of departure for the return flight.

**childrenAges:** The ages of the children specified by the user when the search is made.

**youthAges:** The ages of the youths (12-25) specified by the user when the search is made.

**tripType:** 0 = One way, 1 = Two way trip, 2 = Open-jaw

**domain:** The domain name of the site on which the search was made.

**searchTimestamp:** A UNIX timestamp for when the search was made.

**ticketType:** 0 = Economy, 1 = Bussiness class, 2 = First class, 3 = Economy plus 

**adults:** The number of adults specified in the search.

**flights:** A list of the flights shown to the user. Each item in the list contains:

* **searchIdentifier:** The identifying hash used to tie out clicks to this particular flight.

* **agents:** A list of the OTAs and the price they offered for that flight.

* **trips:** List of the trips such as leave and home trip.

**isBaggageIncluded:** A boolean indicating whether the user has activated the baggage included filter or not.

**legs:** A json list of all searched trip legs. Each leg is a json object that include the following fields:

* **From**: An iata code. Example "ARN".

* **To**: An iata code.

* **Date**: The leave date specified by the user. Example 2018-12-30

The length of this list will be 1 for one way searches, 2 for two way searches and n for open jaw searches (n: number for trip legs). 

**device:** The type of the user device. Could be one of these three values: "DESKTOP", “TABLET” or “MOBILE”.
