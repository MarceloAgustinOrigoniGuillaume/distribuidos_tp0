import socket

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

    def recv_bytes(self, size):
        buf = b""

        while len(buf) < size:
            chunk = self._conn.recv(size - len(buf))
            if not chunk:
                # Connection closed or error
                raise ConnectionError("Socket closed before receiving enough data")
            buf += chunk
        return buf

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


    def close(self):
        self._conn.close()