#!/usr/bin/env python3.6
import socket
import select
import json
import ssl
import sys
from NetProtocol import NetProtocol


class ExampleClient(NetProtocol):

	def __init__(self):
		super().__init__()
		self.token = 'INSERT YOUR TOKEN HERE'
		if self.token == 'INSERT YOUR TOKEN HERE':
			print("You must insert your token in exampleClient.py for the client to be able to connect to the server.")
			sys.exit()
		self.host = 'ai-stream.flightmate.com'
		self.port = 444
		self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
		self.sock.settimeout(20)
		self.context = ssl.create_default_context()
		self.sock = self.context.wrap_socket(self.sock, server_hostname=self.host)

	def _connect(self):
		packet = self._generate_packet(3, self.token.encode('utf-8'))
		self.sock.connect((self.host, self.port))
		self.sock.sendall(packet)

	@staticmethod
	def _handle_search_packet(message):
		print("Got a packet containing a search from %s to %s from %s." % (message['from'], message['to'], message['domain']))

	@staticmethod
	def _handle_click_packet(message):
		print("Got a click packet from %s" % message['domain'])

	def run(self):
		self._connect()
		while True:
			readables, writable, exceptional = select.select([self.sock], [], [])
			for readable in readables:
				header = self._read_header(readable)
				if not header:
					print("Failed to read header.")
					readable.close()
					continue
				
				# Control body size
				body_size = header[1]
				packet_type = header[2]
				if body_size > self.MAX_BODY_SIZE or body_size < self.MIN_BODY_SIZE:
					print("Body size %d, limit is max %d and min %d" % (body_size, self.MAX_BODY_SIZE, self.MIN_BODY_SIZE))
					readable.close()
					continue
				
				body = self._read_socket(readable, body_size, self.READ_SIZE)

				if not self._control_checksum(header, body):
					print("Checksum failed.")
					readable.close()
					continue

				print("Received packet of type: %s" % packet_type)
				message = json.loads(body.decode('utf-8'))
				if packet_type == 1:
					self._handle_search_packet(message)
				elif packet_type == 2:
					self._handle_click_packet(message)
					

if __name__ == "__main__":
	client = ExampleClient()
	client.run()