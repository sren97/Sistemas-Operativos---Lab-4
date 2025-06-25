package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const TAM_MAX_COLA = 10

// ColaSegura implementa una cola thread-safe
type ColaSegura struct {
	elementos []int
	frente    int
	final     int
	tamano    int
	capacidad int
	mutex     sync.Mutex
	no_vacia  *sync.Cond
}

// Inicializa la cola
func inicializar_cola(capacidad int) *ColaSegura {
	cola := &ColaSegura{
		elementos: make([]int, capacidad),
		capacidad: capacidad,
		frente:    0,
		final:     0,
		tamano:    0,
	}
	cola.no_vacia = sync.NewCond(&cola.mutex)
	return cola
}

// Verifica si la cola está vacía
func esta_vacia(cola *ColaSegura) bool {
	return cola.tamano == 0
}

// Verifica si la cola está llena
func esta_llena(cola *ColaSegura) bool {
	return cola.tamano == cola.capacidad
}

// Encola un elemento
func encolar(cola *ColaSegura, elemento int) {
	cola.mutex.Lock()
	defer cola.mutex.Unlock()

	// Espera si la cola está llena
	for esta_llena(cola) {
		fmt.Println("Cola llena, productor esperando...")
		cola.mutex.Unlock()
		time.Sleep(100 * time.Millisecond) // Espera 100ms
		cola.mutex.Lock()
	}

	// Agrega el elemento a la cola
	cola.elementos[cola.final] = elemento
	cola.final = (cola.final + 1) % cola.capacidad
	cola.tamano++

	fmt.Printf("Encolado: %d (Tamaño cola: %d)\n", elemento, cola.tamano)

	// Notifica que la cola ya no está vacía
	cola.no_vacia.Signal()
}

// Desencola un elemento
func desencolar(cola *ColaSegura) int {
	cola.mutex.Lock()
	defer cola.mutex.Unlock()

	// Espera si la cola está vacía
	for esta_vacia(cola) {
		fmt.Println("Cola vacía, consumidor esperando...")
		cola.no_vacia.Wait()
	}

	// Elimina el elemento del frente de la cola
	elemento := cola.elementos[cola.frente]
	cola.frente = (cola.frente + 1) % cola.capacidad
	cola.tamano--

	fmt.Printf("Desencolado: %d (Tamaño cola: %d)\n", elemento, cola.tamano)

	return elemento
}

// Función del hilo productor
func productor(cola *ColaSegura, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= 10; i++ {
		numero := rand.Intn(100)
		encolar(cola, numero)
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond) // Espera aleatoria hasta 500ms
	}

	fmt.Println("Productor finalizó")
}

// Función del hilo consumidor
func consumidor(cola *ColaSegura, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; i <= 5; i++ {
		elemento := desencolar(cola)
		fmt.Printf("Consumidor procesó el elemento: %d\n", elemento)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // Espera aleatoria hasta 1s
	}

	fmt.Println("Consumidor finalizó")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	cola := inicializar_cola(TAM_MAX_COLA)
	var wg sync.WaitGroup

	fmt.Println("Iniciando prueba de cola segura...")

	// Crear hilos productores
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go productor(cola, &wg)
	}

	// Crear hilos consumidores
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go consumidor(cola, &wg)
	}

	// Esperar que todos los hilos terminen
	wg.Wait()

	fmt.Printf("Todos los hilos finalizaron. Tamaño final de la cola: %d\n", cola.tamano)
}
