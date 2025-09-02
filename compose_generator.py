
import sys,os


CLIENT_NAME="Santiago Lionel"
CLIENT_SURNAME="Lorca"
CLIENT_DNI=30904465
CLIENT_BIRTH="1999-03-17"
#CLIENT_NUMERO=7574

def read_base():
	with open("compose_templates/docker-compose-base-dev.yaml", "r") as f:
		return f.read()


def read_client():
	with open("compose_templates/docker-compose-client-dev.yaml", "r") as f:
		return f.read()


# Client config on docker compose instead of .env files to make it easier to configure
def configure_client(base, ind, config_file):
	str_ind = str(ind)
	return base.format(ID = str_ind
	  , CONFIG_FILE=config_file 
	  , INFO_NAME={CLIENT_NAME+"_"+str_ind}
      , INFO_SURNAME={CLIENT_SURNAME}
      , INFO_DNI={CLIENT_DNI+ind}
      , INFO_BIRTH={CLIENT_BIRTH}
      , INFO_NUMBER={ind}
		)


def trim_initial_base(out, base, server_config_file):

	parts = base.split("$CLIENTS", 1) # max split = 1
	out.write(parts[0].format(SERVER_CONFIG_FILE=server_config_file))
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
		base = trim_initial_base(out, base, server_config_file)

		for i in range(1, clients_count+1):
			out.write(configure_client(client_base, i, client_config_file))
		out.write(base)

	print("Finished docker compose creation")

