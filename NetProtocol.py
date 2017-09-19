import struct
from binascii import crc32
from datetime import datetime

DATE_TIME_FORMAT = '%Y-%m-%d, %H:%M:%S'

def time_print(message):
	print("[{}]: {}".format(datetime.now().strftime(DATE_TIME_FORMAT), message))

class NetProtocol(object):
	"""
	This class pack/unpack data that will be sent from and to FlightmateStream
	"""
	def __init__(self):
		super(NetProtocol, self).__init__()

		self.headerFormat = '!LLB'
		self.preHeaderFormat = '!LB'
		self.headerSize = 9
		self.maxBodySize = 2000 * 1024
		self.minBodySize = 1
		self.encoding = 'utf-8'
		self.readSize = 4096
		self.SEARCHPACKET = 1
		self.CLICKPACKET = 2
		self.AUTHPACKET = 3
		self.BADPACKET = -1
	
	def readSocket(self, sock, size, readSize):
		"""
		Reads a certain amount of data from a socket. This method does not close 
		the socket but returns None if the socket breaks before reading the 
		specified amount of data.
		:param socket sock: The socket to read from.
		:param int size: The amount of bytes to read from the socket.
		:return binary: The data read if the specified amount of data was succesfully read.
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
				time_print("Error receivning chunk")
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
			raise ValueError("data read not of correct lengt. Length is %d and should be %d" % (len(data), size))
	
	def readHeader(self, sock):
		"""
		Reads the header of a packet and returns a list of it's content.
		:param socket sock: The socket to read the header from.
		:return list: Returns a list [(int) checksum, (int) bodySize, (int) dataType]. If something goes wrong
		a empty list is returned.
		"""
		try:
			data = self.readSocket(sock, self.headerSize, self.headerSize)
		except (BrokenPipeError, ValueError) as e:
			time_print(e)
			return []
		if data:
			header = struct.unpack(self.headerFormat, data)
		else:
			return []

		# Control header
		if len(header) == 3:
			return header
		return []

	def generatePacket(self, requestType, requestBody):
		"""
		Creates a header and uses it to generate a packet.
		:param int requestType: The type of the request representet by a int.
		:param binary requestBody: The data you want to pack.
		"""
		bodyLength = len(requestBody)
		preHeader = struct.pack(self.preHeaderFormat, bodyLength, requestType)
		message = preHeader + requestBody
		checksum = crc32(message)
		header = struct.pack(self.headerFormat, checksum, bodyLength, requestType)
		return header + requestBody

	def controlChecksum(self, header, body):
		"""
		Controles the checksum of a message.
		:param list header: The header of the message to control. This contains the checksum to check against.
		:param binary body: The body of the message to contorl.
		:return bool: True if the checksums matches otherwise false.
		"""
		checksum = header[0]
		message = struct.pack(self.preHeaderFormat, header[1], header[2]) + body
		return crc32(message) == checksum
