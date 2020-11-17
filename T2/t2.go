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

func send_message_to_address(string_to_send string, address string){

    fmt.Printf("Mandando mensaje: %s \n", string_to_send)
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

    conn.Write([]byte(string_to_send))
}

func send_messages(id int, addr_list []string, instructions []string, queue *[]string, count *int, initializer bool){

    //fmt.Println("LEN INSTRUCTIONS ES")
    //fmt.Println(len(instructions))

    //1 goroutine por instrucción
    breaker:= false
    for ; (!breaker) ; {

        for message_id, order:= range instructions{

            time.Sleep(time.Duration(100) * time.Millisecond)

            //fmt.Printf("message_id es %d \n", message_id)

            //Pre-procesado string de instrucciones
            neworder := strings.Replace(order, "C", "", -1)
            s := strings.Split(neworder, " ")
            i, _ := strconv.Atoi(string(s[0]))

            //SI LA INSTRUCCION ES PARA EL NODO
            if (i == id){
                if s[1] == "M"{

                    fmt.Printf("Incrementando contador antes de mandar mensaje: \n")
                    *count += 1
                    string1_to_send := "MSJ " +strconv.Itoa(message_id) + " " + strconv.Itoa(*count)
                    string2_to_send := "SENDMSJ " + strconv.Itoa(message_id) + " " + strconv.Itoa(*count)
                    *queue = append(*queue, string2_to_send)

                    //CAMBIAR A TODOS LOS DESTINATARIOS
                    var slice []string = s[2:len(s)]
                    for _, slicito:= range slice{
                        i, _ := strconv.Atoi(string(slicito))

                        go send_message_to_address(string1_to_send, addr_list[i])
                        //LO PONGO AL INICIO DE LA QUEUE
                        
                    } 
                } else if (s[1] == "A"){
                    i2, _ := strconv.Atoi(string(s[2]))
                    *count += i2
                    //fmt.Printf("Procesando instruccion:%s\n", order) 
                    fmt.Printf("Self counter += :%d\n", i2) 
                }

                

            }

            if (message_id == len(instructions)-1){
                fmt.Println("LEYENDO FINAL")
                breaker = true
                break
            }else {
                continue
            }
        }
    }
    //MENSAJE DE FIN
    if (initializer){
        //EL PROCESO QUE TERMINA AL NODO DE TERMINO SE DEMORA UN POCO MÁS PARA ASEGURAR QUE PROCESO TERMINE AL RESTO
        
        time.Sleep(20000 * time.Millisecond)
        
        fmt.Println("MANDANDO MENSAJE DE TERMINO A NODOS \n")
        for _, indexito:= range addr_list{
            go send_message_to_address( "FINISH", indexito)
        }

    }
}

func queue_manager(self_id int, addr_list *[]string, queue *[]string, ack_list *[]string, clock *int){
    for{
        time.Sleep(time.Duration(100) * time.Millisecond)       

        //fmt.Println(*queue)
        if len(*queue) != 0{
            time.Sleep(time.Duration(100) * time.Millisecond)

            //fmt.Println("Queue status: \n")
            //fmt.Println(*queue)

            x, _ := pop(*queue)

            //S[1] = id, S[2:] = splice de IDs
            s := strings.Split(x, " ")

            if s[0] == "SENDMSJ"{

                //SI LLEGA ACÁ SOLO ESPERA A QUE LOS ACKS LLEGUEN

                sum := 0

                for ; (sum<len(s)-2); {

                    for _, value := range *ack_list{
                        ack := strings.Split(value, " ")

                        if ack[1] == s[1]{
                            sum += 1
                        }
                    }

                    if (sum == len(s)-2){
                        break
                    }
                    sum = 0
                }

                fmt.Println("Han llegado todos los acks correspondientes a proceso \n")

                

            } else if s[0] == "RECVMSJ"{

                i, _ := strconv.Atoi(s[2])

                if (i>*clock){
                    *clock = i
                    fmt.Printf("INCREMENTANDO RELOJ A : %d\n",i)

                }
                fmt.Println("INCREMENTANDO RELOJ")

                *clock++;

                send_ack_global(self_id, *addr_list, s[1])


            }

            fmt.Printf("Popeando %s \n", (*queue)[0])
            //fmt.Println(queue)
            _, queue2 := pop(*queue)
            *queue = queue2

        }
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

    var queue []string

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
  
		addr, err := net.ResolveUDPAddr("udp4", line)
		if err != nil {
			//fmt.Printf("ERROR: %s\n", err)
			continue
		}
        self_id = index_process
        fmt.Printf("Nodo %d inicializado: \n", self_id)
        fmt.Printf("Set Address: %s \n", line)

        if ((N-1)==self_id && !started){ //SI ES EL ULTIMO NODO DE LA LISTA, ES EL INICIALIZADOR Y LE MANDA A TODOS INIT
            for _, addr := range addresses {
                started = true
                go send_message_to_address("VAMOS", addr)
                
            }
            go send_messages(self_id, addresses, instructions, &queue, &clock, true)


        }
	    // Open up a connection
		//fmt.Println("Open Connection")
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			//fmt.Printf("ERROR: %s %s\n", conn, err)
			continue
		}
		defer conn.Close()

        go queue_manager(self_id, &addresses, &queue, &ack_list, &clock)

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
                go send_messages(self_id, addresses, instructions, &queue, &clock, false)

            } else{

                s := strings.Split(stringer, " ")

                if stringer=="FINISH"{
                    finished = true
                    break
                } else if s[0]== "ACK"{

                    ack_list = append(ack_list, stringer)                    

                } else if s[0]== "MSJ"{

                    stringer = "RECV"+stringer
                    queue = append(queue, stringer)
                    combined_list = append(combined_list, stringer)


                } 

            }
		}

        if (finished == true){
            break
        }

	}
    fmt.Printf("\n")


    fmt.Printf("PROCESO %d TERMINADO\n", self_id)

	fmt.Printf("RELOJ LOGICO: %d \n", clock)

    fmt.Printf("LISTA DE MENSAJES: \n")
    fmt.Println(message_list)

    fmt.Printf("LISTA COMBINADA: \n")
    fmt.Println(combined_list)
    
    fmt.Printf("QUEUE ESTADO FINAL: \n")
    fmt.Println(queue)
	
}
