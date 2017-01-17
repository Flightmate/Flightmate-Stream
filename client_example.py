#!/usr/bin/env python2

import socket, zlib, struct
import ssl

# AUTH PACKETS, THIS IS WHERE YOU WILL PLACE THE KEYS YOU GET FROM FLIGHTMATE
# BELOW ARE JUST SAMPLE KEYS
authPackets = [
	'oUH3P1EVkErSoZ4vfp6oZ6tc5MmYjaC1VeA9gMwkPF2z4BNkAoIggDLJYu7QjnTIWuTAYtik6nADaaF1cKT37obwdT3xIXos4WuJSbTHzo9a1O4rO6ztvGjfnzz5W3xh',
	'n7dQqhNP4X9pWFrou3W5xQMTKpFI7t2EKc6gtVCE2qwsfIYnRyFjjrXs9knMMBYm81x03uxf26ZoaUezSWKqBnfMZjN9LwNvZgO50NKvGfmdBL0jLzkDh6Gihgg22W6e'
	]

# SERVER INFORMATION
HOST = '127.0.0.1'
PORT = 5000

# CREATE SOCKET
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.settimeout(20)

# WRAP SOCKET
wrappedSocket = ssl.wrap_socket(sock, ssl_version=ssl.PROTOCOL_TLSv1_2)

# CONNECT AND PRINT REPLY
wrappedSocket.connect((HOST, PORT))

while True:
	# SEND AUTHENTICATION PACKET(S)
	for authPacket in authPackets:
		wrappedSocket.send(authPacket)

	# GET BYTE LENGTH OF INCOMING PACKET
	# !I = unsigned int with big endian, type size = 4 byte
	dataToRead = struct.unpack("!I", wrappedSocket.read(4))[0]
	data = ""
	data_counter = 0

	# RECEIVE CHUNKS UNTIL dataToRead bytes IS READ
	while data_counter < dataToRead:
		temp_data = wrappedSocket.recv(1024)
		data_counter += len(temp_data)
		data += temp_data
	
	# DECOMPRESS THE DATA
	decompressed = zlib.decompress(data)

	# HERE IS WHERE YOUR PROCESSING WILL START
	print(decompressed)

# CLOSE SOCKET CONNECTION will never happen
wrappedSocket.close()

