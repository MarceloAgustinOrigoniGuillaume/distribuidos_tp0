# TP0: Docker + Comunicaciones + Concurrencia
## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

Previamente se puede correr:
| `./generar-compose.sh <output> <number_of_clients>` | Genera el docker compose a tal output con la cantidad de clientes |

Para el ejercicio 7, esta rama, sera necesario que se pongan el dataset.zip, los csv extraerlos a la carpeta .data . O agregar manualmente .csv validos.

### Servidor

El servidor se mantuvo con 1 solo thread, entendiendo que es a lo que apuntaba el ejercicio, por lo cual no se puede tener un thread a la espera de una signal del signal handler. El servidor recibe las apuestas de cada agencia 1 a la vez. Dejando una conexion persistente con las que ya enviaron sus apuestas, hasta que la ultima lo haga. Una vez eso pase se mandaran los ganadores. De forma stremeada, para evitar guardar en memoria todo ganador.

### Cliente

En el cliente ya se tenia la notificacion de cuando se enviaban todas las apuestas en el ej6. Esto simplemente es mandando un batch ==0. Como siempre se va a querer esperar al sorteo no hace falta mandar ninguna consulta extra y directamente va a esperar a que le llegen los ganadores. Para lo cual recibe primero el numero de la apuesta ganadora, no solo para tener mas info de la apuesta sino para poder mandar -1 como EOF de los ganadores en vez de tener que mandar un flag de continue que sea extra, o de precalcular cuantos ganadores hubo.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.