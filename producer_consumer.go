package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

const (
    bufferSize       = 10
    numProducers     = 2
    numConsumers     = 3
    itemsPerProducer = 8
)

var (
    buffer        [bufferSize]int
    inputIndex    = 0
    outputIndex   = 0
    emptySlots    = make(chan struct{}, bufferSize) // semáforo de espacios vacíos
    fullSlots     = make(chan struct{}, bufferSize) // semáforo de espacios llenos
    bufferMutex   sync.Mutex
    countMutex    sync.Mutex
    totalProduced = 0
    totalConsumed = 0
)

// produceItem simula la producción de un ítem
func produceItem() int {
    return rand.Intn(1000)
}

// consumeItem simula el consumo de un ítem
func consumeItem(item int) {
    fmt.Printf("Consumed: %d\n", item)
    time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
}

// producer coloca items en el buffer usando emptySlots/fullSlots
func producer(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    for i := 0; i < itemsPerProducer; i++ {
        item := produceItem()

        // Espera un token de ranura vacía
        <-emptySlots

        // Sección crítica: insertar en buffer
        bufferMutex.Lock()
        buffer[inputIndex] = item
        fmt.Printf("Producer %d produced %d at position %d\n", id, item, inputIndex)
        inputIndex = (inputIndex + 1) % bufferSize
        bufferMutex.Unlock()

        // Incrementa contador de producidos
        countMutex.Lock()
        totalProduced++
        countMutex.Unlock()

        // Señala ranura llena
        fullSlots <- struct{}{}

        time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
    }
    fmt.Printf("Producer %d finished\n", id)
}

// consumer toma items del buffer o, si ya se consumió todo, sale sin quedarse bloqueado
func consumer(id int, wg *sync.WaitGroup, totalToProduce int) {
    defer wg.Done()

    for {
        // Bloquea hasta que haya un fullSlot disponible
        <-fullSlots

        // Comprueba si ya se han consumido todos los ítems
        countMutex.Lock()
        if totalConsumed >= totalToProduce {
            countMutex.Unlock()
            return // sale sin procesar ningún valor residual
        }
        countMutex.Unlock()

        // Sección crítica: extraer del buffer
        bufferMutex.Lock()
        item := buffer[outputIndex]
        buffer[outputIndex] = 0 // opcional: limpiar celda
        outputIndex = (outputIndex + 1) % bufferSize
        bufferMutex.Unlock()

        // Actualiza contador de consumidos
        countMutex.Lock()
        totalConsumed++
        countMutex.Unlock()

        // Señala ranura vacía
        emptySlots <- struct{}{}

        // Procesa el ítem
        consumeItem(item)
    }
}

// displayBuffer imprime el estado final
func displayBuffer() {
    bufferMutex.Lock()
    defer bufferMutex.Unlock()

    fmt.Println("\n=== Buffer State ===")
    fmt.Printf("Empty slots: %d, full slots: %d\n", bufferSize-len(fullSlots), len(fullSlots))
    fmt.Printf("Total produced: %d, consumed: %d\n", totalProduced, totalConsumed)
    fmt.Print("Content: [")
    for i := 0; i < bufferSize; i++ {
        if i == inputIndex {
            fmt.Print("IN->")
        }
        if i == outputIndex {
            fmt.Print("OUT->")
        }
        fmt.Print(buffer[i])
        if i < bufferSize-1 {
            fmt.Print(", ")
        }
    }
    fmt.Println("]")
    fmt.Println("===================\n")
}

func main() {
    rand.Seed(time.Now().UnixNano())

    totalToProduce := numProducers * itemsPerProducer
    var wgProducers, wgConsumers sync.WaitGroup

    fmt.Println("Producer-Consumer Simulation")
    fmt.Printf("Buffer size: %d\n", bufferSize)
    fmt.Printf("Producers: %d (each produces %d)\n", numProducers, itemsPerProducer)
    fmt.Printf("Consumers: %d\n", numConsumers)
    fmt.Printf("Total to produce: %d items\n\n", totalToProduce)

    // Inicializa semáforos de emptySlots
    for i := 0; i < bufferSize; i++ {
        emptySlots <- struct{}{}
    }

    // Arranca productores
    wgProducers.Add(numProducers)
    for i := 0; i < numProducers; i++ {
        go producer(i+1, &wgProducers)
    }

    // Arranca consumidores
    wgConsumers.Add(numConsumers)
    for i := 0; i < numConsumers; i++ {
        go consumer(i+1, &wgConsumers, totalToProduce)
    }

    // Espera a que todos los productores terminen
    wgProducers.Wait()

    // Luego de producir todo, inyecta tokens adicionales en fullSlots
    // para despertar a consumidores que pudieran estar bloqueados.
    for i := 0; i < numConsumers; i++ {
        fullSlots <- struct{}{}
    }

    // Ahora sí, espera a que todos los consumidores terminen
    wgConsumers.Wait()

    fmt.Println("\nAll threads finished.")
    displayBuffer()
}
