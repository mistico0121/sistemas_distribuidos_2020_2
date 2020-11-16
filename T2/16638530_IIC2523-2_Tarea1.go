package main

import (
    "bufio"   // Leer STDIN
    //"flag"    // Leer flags
    "fmt"     // Print
    "log"     // Escribir Logs
    "os"      // Leer sistema de archivos
    //"runtime" // Imprimir # de GoRouitnes
    //"strings" // Usar Replace()
    "sync"
    "encoding/csv"     // Leer archivos csv
    "strings"

    "time"

    "strconv" // Convertir de string a int
)


func main() {

    string filename1 
    filename1 = os.Args[1]

    //string filename2
    //filename2 = os.Args[1]

    //csvFile, err := os.Open(filename1)
    //if err != nil {
    //    fmt.Println(err)
    //}

    //defer csvFile.Close()

    //r  := csv.NewReader(csvFile)
    //col, err := r.Read()
    //if err != nil {
    //    fmt.Println(err)
    //}
    
    //csvLines, err := r.ReadAll()
    //if err != nil {
    //    fmt.Println(err)
    //}

    fmt.Fprintf("Testing: %s", filename1)
    
    

}


