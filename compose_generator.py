
import sys


def read_base():
	with open("compose_templates/docker-compose-base-dev.yaml", "r") as f:
		return f.read()


def read_client():
	with open("compose_templates/docker-compose-client-dev.yaml", "r") as f:
		return f.read()


def configure_client(base, ind = 1):
	return base.format(ID = str(ind), LOG_LEVEL= "DEBUG")


def trim_initial_base(out, base):

	parts = base.split("$CLIENTS", 1) # max split = 1
	out.write(parts[0])
	return parts[1]

if __name__ == "__main__":

	params = sys.argv

	if len(params) < 3:
		print("> Required two parameters <out_file_name> <number_of_clients>")
		exit()

	file_out = params[1]
	clients_count = 1
	try:
		clients_count = int(params[2])
	except Exception as e:
		print("Invalid number of clients",e)
		exit()

	print(f">Output file:{file_out}, number of clients {clients_count}")

	base = read_base()
	client_base = read_client()
	with open(file_out, "w+") as out:
		base = trim_initial_base(out, base)

		for i in range(1, clients_count+1):
			out.write(configure_client(client_base, i))
		out.write(base)

	print("Finished docker compose creation")

