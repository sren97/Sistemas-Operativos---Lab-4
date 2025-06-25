package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	TAM_BUFFER              = 10
	NUM_PRODUCTORES         = 2
	NUM_CONSUMIDORES        = 3
	ELEMENTOS_POR_PRODUCTOR = 8
)

var (
	buffer           [TAM_BUFFER]int
	entrada          = 0
	salida           = 0
	vacios           = make(chan struct{}, TAM_BUFFER) // semáforo de espacios vacíos
	llenos           = make(chan struct{}, TAM_BUFFER) // semáforo de espacios llenos
	mutex_buffer     sync.Mutex
	mutex_contador   sync.Mutex
	total_producidos = 0
	total_consumidos = 0
)

// Genera un nuevo elemento (simulado)
func producir_elemento() int {
	return rand.Intn(1000)
}

// Procesa un elemento (simulado)
func consumir_elemento(elemento int) {
	fmt.Printf("Consumido: %d\n", elemento)
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
}

// Función del hilo productor
func productor(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < ELEMENTOS_POR_PRODUCTOR; i++ {
		elemento := producir_elemento()

		// Espera un token de ranura vacía
		<-vacios

		// Sección crítica: insertar en buffer
		mutex_buffer.Lock()
		buffer[entrada] = elemento
		fmt.Printf("Productor %d produjo %d en posición %d\n", id, elemento, entrada)
		entrada = (entrada + 1) % TAM_BUFFER
		mutex_buffer.Unlock()

		// Incrementa contador de producidos
		mutex_contador.Lock()
		total_producidos++
		mutex_contador.Unlock()

		// Señala ranura llena
		llenos <- struct{}{}

		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	}
	fmt.Printf("Productor %d finalizó\n", id)
}

// Función del hilo consumidor
func consumidor(id int, wg *sync.WaitGroup, total_a_producir int) {
	defer wg.Done()
	consumidos_local := 0

	for {
		// Espera si no hay elementos
		<-llenos

		// Comprueba si ya se han consumido todos los ítems
		mutex_contador.Lock()
		if total_consumidos >= total_a_producir {
			mutex_contador.Unlock()
			break
		}

		// Sección crítica: extraer del buffer
		mutex_buffer.Lock()
		elemento := buffer[salida]
		buffer[salida] = 0 // Limpia la celda
		salida = (salida + 1) % TAM_BUFFER

		total_consumidos++
		consumidos_local++
		mutex_contador.Unlock()

		mutex_buffer.Unlock()

		// Libera un espacio
		vacios <- struct{}{}

		// Procesa el ítem
		consumir_elemento(elemento)
	}

	fmt.Printf("Consumidor %d finalizó (consumió %d elementos)\n", id, consumidos_local)
}

// Muestra el estado del buffer
func mostrar_estado_buffer() {
	mutex_buffer.Lock()
	defer mutex_buffer.Unlock()

	fmt.Println("\n=== Estado del Buffer ===")
	fmt.Printf("Espacios vacíos: %d, llenos: %d\n", len(vacios), len(llenos))
	fmt.Printf("Total producidos: %d, consumidos: %d\n", total_producidos, total_consumidos)
	fmt.Print("Contenido: [")
	for i := 0; i < TAM_BUFFER; i++ {
		if i == entrada {
			fmt.Print("ENT->")
		}
		if i == salida {
			fmt.Print("SAL->")
		}
		fmt.Print(buffer[i])
		if i < TAM_BUFFER-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println("]")
	fmt.Println("=========================\n")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	total_a_producir := NUM_PRODUCTORES * ELEMENTOS_POR_PRODUCTOR
	var wg_productores, wg_consumidores sync.WaitGroup

	fmt.Println("Simulación Productores-Consumidores")
	fmt.Printf("Tamaño del buffer: %d\n", TAM_BUFFER)
	fmt.Printf("Productores: %d (cada uno produce %d)\n", NUM_PRODUCTORES, ELEMENTOS_POR_PRODUCTOR)
	fmt.Printf("Consumidores: %d\n", NUM_CONSUMIDORES)
	fmt.Printf("Total a producir: %d elementos\n\n", total_a_producir)

	// Inicializa semáforos de vacios
	for i := 0; i < TAM_BUFFER; i++ {
		vacios <- struct{}{}
	}

	// Arranca productores
	wg_productores.Add(NUM_PRODUCTORES)
	for i := 0; i < NUM_PRODUCTORES; i++ {
		go productor(i+1, &wg_productores)
	}

	// Arranca consumidores
	wg_consumidores.Add(NUM_CONSUMIDORES)
	for i := 0; i < NUM_CONSUMIDORES; i++ {
		go consumidor(i+1, &wg_consumidores, total_a_producir)
	}

	// Espera a que todos los productores terminen
	wg_productores.Wait()

	fmt.Println("\nTodos los productores terminaron. Esperando consumidores...")

	// Luego de producir todo, inyecta tokens adicionales en llenos
	// para despertar a consumidores que pudieran estar bloqueados.
	for i := 0; i < NUM_CONSUMIDORES; i++ {
		llenos <- struct{}{}
	}

	// Ahora sí, espera a que todos los consumidores terminen
	wg_consumidores.Wait()

	fmt.Println("\nTodos los hilos terminaron.")
	mostrar_estado_buffer()
}
