#!/usr/bin/env python3.6
import socket, select, time, json
from NetProtocol import NetProtocol

class ExampleClient(NetProtocol):

	def __init__(self):
		super().__init__()
		self.token = 'INSERT YOUR TOKEN HERE'
		self.host = 'ai-stream.flightmate.com'
		self.port = 443
		self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
		self.sock.settimeout(20)

	def _connect(self):
		packet = self.generatePacket(3, self.token.encode('utf-8'))
		self.sock.connect((self.host, self.port))
		self.sock.sendall(packet)

	def _handle_search_packet(self, message):
		print("Got a search packet")

	def _handle_click_packet(self, message):
		print("Got a click packet")

	def run(self):
		self._connect()
		while True:
			readables, writable, exceptional = select.select([self.sock], [], [])
			for readable in readables:
				header = self.readHeader(readable)
				if not header:
					print("Failed to read header.")
					readable.close()
					continue
				
				# Control body size
				bodySize = header[1]
				packetType = header[2]
				if bodySize > self.maxBodySize or bodySize < self.minBodySize:
					print("Body size %d, limit is max %d and min %d" % (bodySize, self.maxBodySize, self.minBodySize))
					readable.close()
					continue
				
				body = self.readSocket(readable, bodySize, self.readSize)

				if not self.controlChecksum(header, body):
					print("Checksum failed.")
					readable.close()
					continue

				print("Received packet of type: %s" % packetType)
				message = json.loads(body.decode('utf-8'))
				if packetType == 1:
					self._handle_search_packet(message)
				elif packetType == 2:
					self._handle_click_packet(message)
					

if __name__ == "__main__":
	client = ExampleClient()
	client.run()