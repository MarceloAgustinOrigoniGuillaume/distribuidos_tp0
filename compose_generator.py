
import sys,os

def read_base():
	with open("compose_templates/docker-compose-base-dev.yaml", "r") as f:
		return f.read()


def read_client():
	with open("compose_templates/docker-compose-client-dev.yaml", "r") as f:
		return f.read()


# Client config on docker compose instead of .env files to make it easier to configure
def configure_client(base, ind, config_file, agency_file):
	str_ind = str(ind)
	return base.format(ID = str_ind
	  , CONFIG_FILE=config_file 
	  , AGENCY_FILE = agency_file
		)


def trim_initial_base(out, base, server_config_file, clients_count):

	parts = base.split("$CLIENTS", 1) # max split = 1
	out.write(parts[0].format(SERVER_CONFIG_FILE=server_config_file, CLIENT_COUNT = clients_count))
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
	ROOT_DIR=os.getcwd()

	server_config_file = os.path.join(ROOT_DIR, "server","config.ini")
	client_config_file = os.path.join(ROOT_DIR, "client","config.yaml")
	print(">Using server config at", server_config_file)
	print(">Using client config at", client_config_file)


	base = read_base()
	client_base = read_client()
	with open(file_out, "w+") as out:
		base = trim_initial_base(out, base, server_config_file, clients_count)

		for i in range(1, clients_count+1):
			client_agency_file = os.path.join(ROOT_DIR, ".data",f"agency-{i}.csv")

			out.write(configure_client(client_base, i, client_config_file, client_agency_file))
		out.write(base)

	print("Finished docker compose creation")

