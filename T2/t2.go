package main

import (
    "bufio"   // Leer STDIN
    //"flag"    // Leer flags
    "fmt"     // Print
    "log"     // Escribir Logs
    "os" 
    "net"    // Leer sistema de archivos
    //"runtime" // Imprimir # de GoRouitnes
    //"strings" // Usar Replace()
    //"sync"
    //"encoding/csv"     // Leer archivos csv
    "encoding/hex"
    "strings"
    "strconv"
    "time"
    "math/rand"

)

const (
	maxDatagramSize = 8192
)

func pop(a []string) (string, []string){
    x, b := a[0], a[1:]

    return x, b
}


// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

func send_ack(self_id int, message_id string, address string) {
    addr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil {
        log.Fatal(err)
    }

    // Dial connects to the address on the named network.
    conn, err := net.DialUDP("udp4", nil, addr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    conn.Write([]byte("ACK " + message_id + " " + strconv.Itoa(self_id)))    
}


func send_ack_global(self_id int, addr_list []string, message_id string){

    for _, addr := range addr_list {
        go send_ack(self_id, message_id, addr)
    }
}


//REPRESENTA FUNCION QUE ENVIA MENSAJES A NODO[addr_id]
func send_messages(id int, addr_id int, address string, instructions []string, count *int,synchro *bool, initializer bool){
    addr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil {
        log.Fatal(err)
    }

    // Dial connects to the address on the named network.
    conn, err := net.DialUDP("udp4", nil, addr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    //MENSAJE DE INICIO
    if (initializer){
        time.Sleep(500 * time.Millisecond)
        fmt.Printf("INICIALIZANDO NODO %d\n", addr_id)
        conn.Write([]byte("VAMOS"))
    }
    

    for message_id, order:= range instructions{
        //fmt.Printf("Procesando instruccion: %s\n", order)
        neworder := strings.Replace(order, "C", "", -1)
        s := strings.Split(neworder, " ")

        i, _ := strconv.Atoi(string(s[0]))

        //SI LA INSTRUCCION ES PARA EL NODO
        if ((i == id) && (id != addr_id)){
            if s[1] == "M"{
                //CAMBIAR A TODOS LOS DESTINATARIOS
                var slice []string = s[2:len(s)]
                for _, receiver:= range slice{

                    ss,_ := strconv.Atoi(receiver)
                    // Dado que 1+ procesos estan intentando incrementar cuando hay un multicast, este pequeño delay
                    // permite que un proceso entre, incremente el contador, y cambie la variable antes de que otro proceso
                    // lo vea, entre al if, y tenga la oportunidad de cambiar el contador mas de 1 vez
                    time.Sleep(time.Duration(rand.Intn(1000)) * time.Microsecond)
                    if (addr_id==ss) {
                        if (*synchro){
                            *synchro = false

                            //fmt.Printf("Incrementando contador antes de mandar mensaje\n")
                            *count += 1
                        }
                        //fmt.Printf("Procesando instruccion: %s\n", order)
                        conn.Write([]byte("MSJ " +strconv.Itoa(message_id) + " " + strconv.Itoa(*count)))
                    }

                }

                
            }
        } else if ((i == id) && (id == addr_id)){
            if (s[1]== "A"){
                i2, _ := strconv.Atoi(string(s[2]))
                *count += i2
                //fmt.Printf("Procesando instruccion:%s\n", order) 
                fmt.Printf("Self counter += :%d\n", i2) 
            }
        }
        

        //Este tiempo representa la diferencia entre cada mensaje que se manda
        time.Sleep(1 * time.Millisecond)

        //puede volver a incrementar el contador usado en multicasts por cada proceso
        *synchro = true
    }

    //MENSAJE DE FIN
    if (initializer){
        //EL PROCESO QUE TERMINA AL NODO DE TERMINO SE DEMORA UN POCO MÁS PARA ASEGURAR QUE PROCESO TERMINE AL RESTO
        if(id ==addr_id ){
            time.Sleep(500 * time.Millisecond)
        }
        time.Sleep(500 * time.Millisecond)
        fmt.Printf("MANDANDO MENSAJE DE TERMINO A NODO %d\n", addr_id)
        conn.Write([]byte("FINISH"))
    }

}



// BASADO EN AYUDANTÍA Y: 
// https://ops.tips/blog/udp-client-and-server-in-go/
func main() {

    var filename1 string
    filename1 = os.Args[1]

    var filename2 string
    filename2 = os.Args[2]

    file, err := os.Open(filename1)
    if err != nil {
        fmt.Println(err)
    }

    defer file.Close()

    // Atributos de cada proceso
    var clock int
    var self_id int

    //PARTIDA Y TERMINO, INDICAN SI SE ESTÁN RECIBIENDO MENSAJES, Y SE USA PARA TERMINAR LA FUNCIÓN
    var started bool
    started = false
    var finished bool

    // Indica si el contador puede incrementar, ya que solo debe incrementar 1 vez por multicast, pero cada multicast
    // Lo ven hasta N-1 procesos
    var synchro bool
    synchro = true

    var message_list []string
    var ack_list []string

    var combined_list []string

    addresses, err := readLines(filename1)
    if err != nil {
        log.Fatalf("readLines: %s", err)
    }

    var newline string
    for i, line := range addresses {
        newline = strings.Replace(line, " ", ":", 1)
        newline = strings.Replace(newline, "\n", "", 1)

        addresses[i] = newline
    }

    instructions, err := readLines(filename2)
    if err != nil {
        log.Fatalf("readLines: %s", err)
    }
    

    //NUMERO DE PUERTOS
    N := len(addresses)
    fmt.Printf("NUMERO DE LINEAS:%d\n", N)
    
    for index_process, line:= range addresses{
  
		//fmt.Println("Set Address: %s \n", line)
		//newline = "127.0.0.1:3000"
		    

		addr, err := net.ResolveUDPAddr("udp4", line)
		if err != nil {
			//fmt.Printf("ERROR: %s\n", err)
			continue
		}
        self_id = index_process
        fmt.Printf("Nodo %d inicializado: \n", self_id)
        fmt.Printf("Set Address: %s \n", line)

        if ((N-1)==self_id && !started){ //SI ES EL ULTIMO NODO DE LA LISTA, ES EL INICIALIZADOR Y LE MANDA A TODOS INIT
            started = true
            for index2, addr := range addresses {
                if (index2!= self_id){
                    go send_messages(self_id, index2, addr, instructions, &clock, &synchro, true)

                }
            }
            go send_messages(self_id, self_id, addresses[self_id], instructions, &clock, &synchro, true)

        }
	    // Open up a connection
		//fmt.Println("Open Connection")
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			//fmt.Printf("ERROR: %s %s\n", conn, err)
			continue
		}
		defer conn.Close()

		fmt.Println("Listening")
		for {
			buffer := make([]byte, maxDatagramSize)
			numBytes, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Fatal("ReadFromUDP failed:", err)
			}

			var string111 string 
            string111 = hex.EncodeToString(buffer[:numBytes])
            bs, _ := hex.DecodeString(string111)
            stringer := string(bs)
            if (stringer=="VAMOS" && !started){
                //CUANDO LLEGA MENSAJE QUE ULTIMO RECEPTOR SE ENCUENTRA LISTO, SE INICIALIZAN SENDERS PARA TODOS LOS OTROS NODOS
                started = true
                for index, addr := range addresses {
                    if (index!= self_id){
                        go send_messages(self_id, index, addr, instructions, &clock , &synchro, false)
                    }
                }
                go send_messages(self_id, self_id, addresses[self_id], instructions, &clock, &synchro, false)


            } else{
                s := strings.Split(stringer, " ")

                if stringer=="FINISH"{
                    finished = true
                    break
                } else if s[0]== "ACK"{

                    ack_list = append(ack_list, stringer)
                    
                    combined_list = append(combined_list, stringer)

                } else if s[0]== "MSJ"{
                    i, _ := strconv.Atoi(s[2])
                    fmt.Printf("RECIBIDO: %s\n",stringer)
                    if (i>clock){
                        clock = i
                    }
                    clock++;

                    fmt.Println("AGREGANDO A LA LISTA DE MENSAJES \n")
                    message_list = append(message_list, stringer)
                    combined_list = append(combined_list, stringer)


                    send_ack_global(self_id, addresses, s[1])

                } 

            }
		}

        if (finished == true){
            break
        }

	}
    
    fmt.Printf("PROCESO %d TERMINADO\n", self_id)

	fmt.Printf("RELOJ LOGICO: %d \n", clock)

    fmt.Printf("LISTA DE MENSAJES: \n")
    fmt.Println(message_list)

    fmt.Printf("LISTA DE ACKS: \n")
    fmt.Println(ack_list)

    fmt.Printf("LISTA COMBINADA: \n")
    fmt.Println(combined_list)
    
	
}
