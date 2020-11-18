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
    millisecondExecution = 10000
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
        time.Sleep(time.Duration(1) * time.Millisecond)
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

func send_messages(id int, addr_list []string, instructions []string, queue *[]string, action_list *[]string ,count *int, initializer bool){

    //1 goroutine por instrucción
    breaker:= false
    for ; (!breaker) ; {

        for message_id, order:= range instructions{

            time.Sleep(time.Duration(100) * time.Millisecond)

            //Pre-procesado string de instrucciones
            neworder := strings.Replace(order, "C", "", -1)
            s := strings.Split(neworder, " ")
            i, _ := strconv.Atoi(string(s[0]))

            //SI LA INSTRUCCION ES PARA EL NODO
            if (i == id){
                if s[1] == "M"{

                    *count += 1
                    string1_to_send := "MSJ " +strconv.Itoa(message_id) + " " + strconv.Itoa(*count)
                    fmt.Printf("Incrementando contador antes de mandar mensaje: %s \n", string1_to_send)
                    string2_to_send := "SENDMSJ " + strconv.Itoa(message_id) + " " + strconv.Itoa(*count)
                    *queue = append(*queue, string2_to_send)
                    *action_list = append(*action_list, string2_to_send)

                    //CAMBIAR A TODOS LOS DESTINATARIOS
                    var slice []string = s[2:len(s)]
                    for _, slicito:= range slice{
                        i, _ := strconv.Atoi(string(slicito))

                        go send_message_to_address(string1_to_send, addr_list[i])
                        //LO PONGO AL INICIO DE LA QUEUE
                        
                    } 
                } else if (s[1] == "A"){

                    string1_to_send := "INCREASE " + string(s[2])

                    *queue = append(*queue, string1_to_send)
                    *action_list = append(*action_list, string1_to_send)
                }

                

            }

            if (message_id == len(instructions)-1){
                //fmt.Println("LEYENDO FINAL")
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
        
        //10 segundos de buffer para terminar
        time.Sleep(millisecondExecution * time.Millisecond)
        
        fmt.Println("MANDANDO MENSAJE DE TERMINO A NODOS")
        for _, indexito:= range addr_list{
            time.Sleep(time.Duration(10) * time.Millisecond)
            go send_message_to_address("FINISH", indexito)
        }

    }
}

func queue_manager(self_id int, addr_list *[]string, queue *[]string, ack_list *[]string, clock *int){
    for{
        //time.Sleep(time.Duration(100) * time.Millisecond)       

        if len(*queue) != 0{
            time.Sleep(time.Duration(50) * time.Millisecond)

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

                fmt.Printf("Han llegado todos los acks correspondientes a proceso %s \n", s[1])
                fmt.Printf("Popeando %s\n", (*queue)[0])
                //fmt.Println(queue)
                _, queue2 := pop(*queue)
                *queue = queue2

                

            } else if s[0] == "RECVMSJ"{

                i, _ := strconv.Atoi(s[2])

                if (i>*clock){
                    *clock = i
                    fmt.Printf("INCREMENTANDO RELOJ A : %d\n",i)

                }
                fmt.Println("INCREMENTANDO RELOJ")

                *clock++;

                send_ack_global(self_id, *addr_list, s[1])
                fmt.Printf("Popeando %s \n", (*queue)[0])
                //fmt.Println(queue)
                _, queue2 := pop(*queue)
                *queue = queue2

            } else if s[0] == "INCREASE"{

                i2, _ := strconv.Atoi(string(s[1]))
                fmt.Printf("Self counter += :%d\n", i2) 
                *clock += i2

                _, queue2 := pop(*queue)
                *queue = queue2

            }

            

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
    
    for index_process, line:= range addresses{
  
		addr, err := net.ResolveUDPAddr("udp4", line)
		if err != nil {
			//fmt.Printf("ERROR: %s\n", err)
			continue
		}
        self_id = index_process
        

        if ((N-1)==self_id && !started){ //SI ES EL ULTIMO NODO DE LA LISTA, ES EL INICIALIZADOR Y LE MANDA A TODOS INIT
            for _, addr := range addresses {
                started = true
                go send_message_to_address("VAMOS", addr)
                
            }
            go send_messages(self_id, addresses, instructions, &queue, &combined_list, &clock, true)


        }
	    // Open up a connection
		//fmt.Println("Open Connection")
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			//fmt.Printf("ERROR: %s %s\n", conn, err)
			continue
		}
		defer conn.Close()

        fmt.Printf("Nodo %d inicializado: \n", self_id)
        fmt.Printf("Set Address: %s \n", line)

        go queue_manager(self_id, &addresses, &queue, &ack_list, &clock)

		fmt.Println("Listening\n")
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
                go send_messages(self_id, addresses, instructions, &queue, &combined_list, &clock, false)

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

    fmt.Printf("QUEUE FINAL: \n")
    output_queue := "'"+strings.Join(queue, `','`) + `'`

    fmt.Println(output_queue)

    fmt.Printf("LISTA MENSAJES ENVIADOS Y RECIBIDOS: \n")

    // prepend single quote, perform joins, append single quote
    output_combined := "'"+strings.Join(combined_list, `','`) + `'`
    fmt.Println(output_combined)
    

	
}
