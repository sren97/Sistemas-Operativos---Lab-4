package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	NUM_FILOSOFOS = 5
	TIEMPO_PENSAR = 2000
	TIEMPO_COMER  = 1000
)

const (
	PENSANDO = iota
	HAMBRIENTO
	COMIENDO
)

var (
	tenedores          [NUM_FILOSOFOS]sync.Mutex
	estado             [NUM_FILOSOFOS]int
	mutex_estado       sync.Mutex
	comensales_comiendo = 0
	limite_comensales  = make(chan struct{}, NUM_FILOSOFOS-1)
)

// Muestra los estados de los filósofos
func mostrar_estado(id int, accion string) {
	mutex_estado.Lock()
	defer mutex_estado.Unlock()

	fmt.Printf("Filósofo %d está %s | Estados: [", id, accion)
	for i := 0; i < NUM_FILOSOFOS; i++ {
		switch estado[i] {
		case PENSANDO:
			fmt.Print("P")
		case HAMBRIENTO:
			fmt.Print("H")
		case COMIENDO:
			fmt.Print("C")
		}
		if i < NUM_FILOSOFOS-1 {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] Comiendo: %d\n", comensales_comiendo)
}

// Índice del tenedor izquierdo
func izquierdo(id int) int {
	return id
}

// Índice del tenedor derecho
func derecho(id int) int {
	return (id + 1) % NUM_FILOSOFOS
}

// Solución 1: Enfoque asimétrico
func tomar_tenedores_asimetrico(id int) {
	izq := izquierdo(id)
	der := derecho(id)

	mutex_estado.Lock()
	estado[id] = HAMBRIENTO
	mutex_estado.Unlock()

	mostrar_estado(id, "hambriento")

	// Los filósofos pares toman primero el izquierdo, los impares el derecho
	if id%2 == 0 {
		tenedores[izq].Lock()
		tenedores[der].Lock()
	} else {
		tenedores[der].Lock()
		tenedores[izq].Lock()
	}

	mutex_estado.Lock()
	estado[id] = COMIENDO
	comensales_comiendo++
	mutex_estado.Unlock()

	mostrar_estado(id, "comiendo")
}

func soltar_tenedores_asimetrico(id int) {
	izq := izquierdo(id)
	der := derecho(id)

	mutex_estado.Lock()
	estado[id] = PENSANDO
	comensales_comiendo--
	mutex_estado.Unlock()

	tenedores[izq].Unlock()
	tenedores[der].Unlock()

	mostrar_estado(id, "pensando")
}

// Solución 2: Enfoque con semáforo
func tomar_tenedores_semaforo(id int) {
	mutex_estado.Lock()
	estado[id] = HAMBRIENTO
	mutex_estado.Unlock()

	mostrar_estado(id, "hambriento")

	// Espera por semáforo - limita los comensales concurrentes
	limite_comensales <- struct{}{}

	izq := izquierdo(id)
	der := derecho(id)

	// Siempre toma los tenedores en un orden consistente para evitar interbloqueos
	if izq < der {
		tenedores[izq].Lock()
		tenedores[der].Lock()
	} else {
		tenedores[der].Lock()
		tenedores[izq].Lock()
	}

	mutex_estado.Lock()
	estado[id] = COMIENDO
	comensales_comiendo++
	mutex_estado.Unlock()

	mostrar_estado(id, "comiendo")
}

func soltar_tenedores_semaforo(id int) {
	izq := izquierdo(id)
	der := derecho(id)

	mutex_estado.Lock()
	estado[id] = PENSANDO
	comensales_comiendo--
	mutex_estado.Unlock()

	tenedores[izq].Unlock()
	tenedores[der].Unlock()

	<-limite_comensales // Libera el semáforo

	mostrar_estado(id, "pensando")
}

// Filósofo con solución asimétrica
func filosofo_asimetrico(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 3; i++ {
		mostrar_estado(id, "pensando")
		time.Sleep(time.Duration(rand.Intn(TIEMPO_PENSAR)) * time.Millisecond)

		tomar_tenedores_asimetrico(id)
		time.Sleep(time.Duration(rand.Intn(TIEMPO_COMER)) * time.Millisecond)
		soltar_tenedores_asimetrico(id)
	}

	fmt.Printf("Filósofo %d terminó de comer\n", id)
}

// Filósofo con solución de semáforo
func filosofo_semaforo(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 3; i++ {
		mostrar_estado(id, "pensando")
		time.Sleep(time.Duration(rand.Intn(TIEMPO_PENSAR)) * time.Millisecond)

		tomar_tenedores_semaforo(id)
		time.Sleep(time.Duration(rand.Intn(TIEMPO_COMER)) * time.Millisecond)
		soltar_tenedores_semaforo(id)
	}

	fmt.Printf("Filósofo %d terminó de comer\n", id)
}

func ejecutar_asimetrico() {
	var wg sync.WaitGroup

	fmt.Println("\n=== Solución con Orden Asimétrico ===")

	for i := 0; i < NUM_FILOSOFOS; i++ {
		estado[i] = PENSANDO
	}
	comensales_comiendo = 0

	for i := 0; i < NUM_FILOSOFOS; i++ {
		wg.Add(1)
		go filosofo_asimetrico(i, &wg)
	}

	wg.Wait()
	fmt.Println("Solución asimétrica finalizada")
}

func ejecutar_semaforo() {
	var wg sync.WaitGroup

	fmt.Println("\n=== Solución con Semáforo ===")

	for i := 0; i < NUM_FILOSOFOS; i++ {
		estado[i] = PENSANDO
	}
	comensales_comiendo = 0

	for i := 0; i < NUM_FILOSOFOS; i++ {
		wg.Add(1)
		go filosofo_semaforo(i, &wg)
	}

	wg.Wait()
	fmt.Println("Solución con semáforo finalizada")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Problema de los Filósofos Comensales")
	fmt.Printf("Filósofos: %d | Cada uno comerá 3 veces\n", NUM_FILOSOFOS)

	ejecutar_asimetrico()
	time.Sleep(2 * time.Second)
	ejecutar_semaforo()
}