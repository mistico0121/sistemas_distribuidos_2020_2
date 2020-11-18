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

    //fmt.Printf("Mandando mensaje: %s \n", string_to_send)
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

func send_messages(id int, addr_list []string, instructions []string, queue *[]string, action_list *[]string ,count *[]int, initializer bool){

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
                //SOLO TOMAMOS EN CUENTA "M", YA QUE "A" NO VALE PARA RELOJES VECTORIALES
                if s[1] == "M"{

                    old_count := *count
                    string_old_count := fmt.Sprintf("Reloj vectorial viejo: %d", old_count)

                    (*count)[id] += 1
                    vector_as_string := []string{}

                    for i := range (*count) {
                        number := (*count)[i]
                        text := strconv.Itoa(number)
                        vector_as_string = append(vector_as_string, text)
                    }

                    // Join our string slice.
                    result := strings.Join(vector_as_string, " ")


                    string1_to_send := "MSJ " +strconv.Itoa(message_id) + " " + result
                    fmt.Println("Procediendo a mandar mensaje")
                    fmt.Println(string_old_count)
                    fmt.Printf("Reloj vectorial nuevo: %d\n", (*count))

                    *action_list = append(*action_list, string1_to_send)

                    //ENVIAR A TODOS LOS DESTINATARIOS
                    var slice []string = s[2:len(s)]
                    for _, slicito:= range slice{
                        i, _ := strconv.Atoi(string(slicito))
                        fmt.Printf("Mandando mensaje: %s a proceso ID: %d \n", string1_to_send, i)

                        go send_message_to_address(string1_to_send, addr_list[i])
                        
                    }
                    fmt.Println("\n") 
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

func queue_manager(self_id int, addr_list *[]string, queue *[]string, ack_list *[]string, clock *[]int){
    for{
        //time.Sleep(time.Duration(100) * time.Millisecond)       

        if len(*queue) != 0{
            time.Sleep(time.Duration(50) * time.Millisecond)

            //checkeo toda la queue
            for _, x := range (*queue){

                //S[1] = id, S[2:] = splice de IDs
                s := strings.Split(x, " ")

                if s[0] == "RECVMSJ"{

                    slicito := s[2:]

                    difference_counter := 0
                    difference_index := 0

                    for indexcito, itemito := range slicito{

                        i, _ := strconv.Atoi(string(itemito))

                        if i > (*clock)[indexcito]{
                            difference_counter++
                            difference_index = indexcito
                        }

                    }

                    //SI SOLO HAY 1 NUMERO EN EL VECTOR QUE LLEGO QUE TENGA UN CAMBIO QUE ESTE NO TIENE
                    if difference_counter == 1{
                        i, _ := strconv.Atoi(string(slicito[difference_index]))

                        fmt.Printf("INCREMENTANDO RELOJ SEGUN MENSAJE ID %s\n", s[1])
                        fmt.Printf("VIEJO RELOJ: %d\n", *clock)
                        
                        (*clock)[difference_index] = i

                        fmt.Printf("NUEVO RELOJ: %d\n", *clock)

                        fmt.Printf("Popeando %s \n", (*queue)[0])
                        //fmt.Println(queue)
                        _, queue2 := pop(*queue)
                        *queue = queue2
                        fmt.Println("\n")

                    }

                    

                }

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
    var self_id int

    //PARTIDA Y TERMINO, INDICAN SI SE ESTÁN RECIBIENDO MENSAJES, Y SE USA PARA TERMINAR LA FUNCIÓN
    var started bool
    started = false
    var finished bool

    var queue []string
    var ack_list []string
    var combined_list []string

    var vector_clock []int

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

    for i:= 0; i< len(addresses); i++ {
        vector_clock = append(vector_clock, 0)
    }
    
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
            go send_messages(self_id, addresses, instructions, &queue, &combined_list, &vector_clock, true)


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

        go queue_manager(self_id, &addresses, &queue, &ack_list, &vector_clock)

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
                go send_messages(self_id, addresses, instructions, &queue, &combined_list, &vector_clock, false)

            } else{

                s := strings.Split(stringer, " ")

                if stringer=="FINISH"{
                    finished = true
                    break
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

	fmt.Printf("RELOJ VECTORIAL: %d \n", vector_clock)

    fmt.Printf("QUEUE FINAL: \n")
    output_queue := "'"+strings.Join(queue, `','`) + `'`
    fmt.Println(output_queue)

    fmt.Printf("LISTA MENSAJES ENVIADOS Y RECIBIDOS: \n")

    // prepend single quote, perform joins, append single quote
    output_combined := strings.Join(combined_list, `' '`)
    fmt.Println(output_combined)
    

	
}
