#!/usr/bin/env python3.6

import socket, select
import struct
from binascii import crc32

def readSocket(sock, size, readSize):
		"""
		Reads a certain amount of data from a socket. This function does not close 
		the socket but returns None if the socket breaks before reading the 
		specified amount of data.
		:param socket sock: The socket to read from.
		:param int size: The amount of bytes to read from the socket.
		:return binary: The data read if the specified amount of data was succesfully 
		read otherwise None.
		"""
		totRead = 0
		chunks = []
		while totRead < size:
			sizeLeft = size - totRead
			if sizeLeft < readSize:
				readSize = sizeLeft
			try:
				chunk = sock.recv(readSize)
			except:
				print("Error receivning chunk")
				return None
			chunkSize = len(chunk)
			if chunkSize == 0:
				raise BrokenPipeError("The socket connection has been broken.")
			totRead += chunkSize
			chunks.append(chunk)
		data = b''.join(chunks)
		if len(data) == size:
			return data
		else:
			print("data read not of correct lengt. Length is %d and should be %d" % (len(data), size))
			return None

TOKEN = '9SEB3boMu5KpFRTj5cncYUNmvVAd35xfsneynrfQhkTJYO1ub7zReuYShKaOr7Ougxaa3vJVfuU0zhDIVMbn6IK6EzEqqNlz2reS0gusSv0d85nv3YFaXC5z5NNT2jif' # Replace with your secret 128 bit long token.
HOST = 'ai-stream.flightmate.com'
PORT = 443

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.settimeout(20)

try:
	# Connect to the stream server
	sock.connect((HOST, PORT))

	# Create the authentication packet
	body = TOKEN.encode('utf-8')
	bodyLength = len(body)
	preHeader = struct.pack('!LB', bodyLength, 3)
	message = preHeader + body
	checksum = crc32(message)
	header = struct.pack('!LLB', checksum, bodyLength, 3)
	authenticationPacket = header + body

	# Send the authentication packet
	sock.sendall(authenticationPacket)

	# Receive data
	running = True
	while running:
		readables, writable, exceptional = select.select([sock], [], [])
		for readable in readables:
			# Read header
			headerData = readSocket(sock, 9, 9)
			if not headerData:
				continue
			header = struct.unpack('!LLB', headerData)

			# Read body
			bodySize = header[1]
			body = readSocket(sock, bodySize, 4096)

			# Control checksum
			checksum = header[0]
			message = struct.pack('!LB', header[1], header[2]) + body
			if crc32(message) != checksum:
				print("Checksum failed")
				continue;

			print(body.decode('utf-8'))

finally:
	sock.close()