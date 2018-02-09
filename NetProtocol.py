import struct
import logging
from binascii import crc32


class StreamError(Exception):
	pass


class NetProtocol(object):
	"""
	This class pack/unpack data that will be sent from and to FlightmateStream
	"""

	def __init__(self):
		super(NetProtocol, self).__init__()
		self.HEADER_FORMAT = '!LLB'
		self.PRE_HEADER_FORMAT = '!LB'
		self.HEADER_SIZE = 9
		self.MAX_BODY_SIZE = 2000 * 1024
		self.MIN_BODY_SIZE = 1
		self.ENCODING = 'utf-8'
		self.READ_SIZE = 4096
		self.BAD_PACKET = -1
		self.SEARCH_PACKET = 1
		self.CLICK_PACKET = 2
		self.OTA_AUTH_PACKET = 3
		self.MASKED_PACKET = 4
		self.POSTBOY_AUTH_PACKET = 5

	@staticmethod
	def _read_socket(sock, size, read_size):
		"""
		Reads a certain amount of data from a socket. This method does not close
		the socket but returns None if the socket breaks before reading the
		specified amount of data.
		:param socket.Socket sock: The socket to read from.
		:param int size: The amount of bytes to read from the socket.
		:param int read_size: The total amount of bytes to read.
		:return binary: The data read if the specified amount of data was succesfully read.
		"""
		tot_read = 0
		chunks = []
		while tot_read < size:
			size_left = size - tot_read
			if size_left < read_size:
				read_size = size_left
			try:
				chunk = sock.recv(read_size)
			except Exception:
				logging.exception("Error receiving chunk")
				raise
			chunk_size = len(chunk)
			if chunk_size == 0:
				raise BrokenPipeError("The socket connection has been broken.")
			tot_read += chunk_size
			chunks.append(chunk)
		data = b''.join(chunks)
		if len(data) == size:
			return data
		else:
			raise ValueError("Data read not of correct length. Length is %d and should be %d" % (len(data), size))

	def _read_header(self, sock):
		"""
		Reads the header of a packet and returns a list of it's content.
		:param socket sock: The socket to read the header from.
		:return list: Returns a list [(int) checksum, (int) bodySize, (int) dataType]. If something goes wrong
		a empty list is returned.
		"""
		data = self._read_socket(sock, self.HEADER_SIZE, self.HEADER_SIZE)
		if data:
			header = struct.unpack(self.HEADER_FORMAT, data)
		else:
			raise StreamError("No data to unpack while trying to read the packet header.")

		# Control header
		if len(header) == 3:
			return header
		raise StreamError("Unpacked header of length %d instead of 3", len(header))

	def _generate_packet(self, request_type, request_body):
		"""
		Creates a header and uses it to generate a packet.
		:param int request_type: The type of the request representet by a int.
		:param binary request_body: The data you want to pack.
		"""
		body_length = len(request_body)
		pre_header = struct.pack(self.PRE_HEADER_FORMAT, body_length, request_type)
		message = pre_header + request_body
		checksum = crc32(message)
		header = struct.pack(self.HEADER_FORMAT, checksum, body_length, request_type)
		return header + request_body

	def _control_checksum(self, header, body):
		"""
		Controls the checksum of a message.
		:param tuple header: The header of the message to control. This contains the checksum to check against.
		:param binary body: The body of the message to contorl.
		:return bool: True if the checksums matches otherwise false.
		"""
		checksum = header[0]
		message = struct.pack(self.PRE_HEADER_FORMAT, header[1], header[2]) + body
		return crc32(message) == checksum

	def _read_packet(self, sock):
		"""
		Reads a packet from a socket and returns the packet body and it's type. Raises a StreamError upon failure.
		:param socket sock: The socket to read the packet from.
		:return (int, string): The type of the packet and it's content.
		"""
		# Read the header.
		header = self._read_header(sock)
		# Control body size
		body_size = header[1]
		if body_size > self.MAX_BODY_SIZE or body_size < self.MIN_BODY_SIZE:
			raise StreamError(
				"Body size %d, limit is max %d and min %d" % (body_size, self.MAX_BODY_SIZE, self.MIN_BODY_SIZE))

		body = self._read_socket(sock, body_size, self.READ_SIZE)

		if not self._control_checksum(header, body):
			raise StreamError("Checksum failed.")

		packet_type = header[2]
		return packet_type, body
