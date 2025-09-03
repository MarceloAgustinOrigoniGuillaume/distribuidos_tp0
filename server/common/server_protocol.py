import socket

TIMEOUT_SECONDS = 3
class BetDTO:
    def __init__(self, first_name: str, last_name: str, document: int, birthdate: str, number: int):
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = birthdate
        self.number = number


class ServerProtocol:
    def __init__(self, conn: socket.socket):
        self._conn = conn
        conn.settimeout(TIMEOUT_SECONDS)

    def _recv_bytes(self, size):
        buf = b""

        while len(buf) < size:
            chunk = self._conn.recv(size - len(buf))
            if not chunk:
                # Connection closed or error
                raise ConnectionError("Socket closed before receiving enough data")
            buf += chunk
        return buf

    def _send_bytes(self, data: bytes):
        total_sent = 0
        while total_sent < len(data):
            sent = self._conn.send(data[total_sent:])
            if sent == 0:
                raise ConnectionError("Socket connection broken during send")
            total_sent += sent



    def _recv_int32(self):
        uint_bytes = self._recv_bytes(4)
        value = int.from_bytes(uint_bytes, byteorder='big', signed=True)
        return value
    def _recv_u16(self):
        uint_bytes = self._recv_bytes(2)
        value = int.from_bytes(uint_bytes, byteorder='big', signed=False)
        return value

    def recv_str(self):
        len_str = self._recv_u16()
        str_bytes = self._recv_bytes(len_str)
        return str_bytes.decode("utf-8")

    def recv_bet(self) -> BetDTO:
        name = self.recv_str()
        surname = self.recv_str()
        dni = self._recv_int32()
        birth = self.recv_str()
        number = self._recv_int32()

        return BetDTO(name, surname, dni, birth, number)




    def send_int32(self, value: int):
        int_bytes = value.to_bytes(4, byteorder='big', signed=True)
        self._send_bytes(int_bytes)

    def _send_u16(self, value: int):
        u16_bytes = value.to_bytes(2, byteorder='big', signed=False)
        self._send_bytes(u16_bytes)

    def send_str(self, value: str):
        encoded = value.encode("utf-8")
        self._send_u16(len(encoded))  # send length first
        self._send_bytes(encoded)


    def close(self):
        self._conn.close()