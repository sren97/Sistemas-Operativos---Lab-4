#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <unistd.h>
#include <time.h>

#define TAM_MAX_COLA 10

typedef struct
{
    int *elementos;
    int frente, final, tamano, capacidad;
    pthread_mutex_t mutex;
    pthread_cond_t no_vacia;
} ColaSegura;

// Inicializa la cola
void inicializar_cola(ColaSegura *cola, int capacidad)
{
    cola->elementos = (int *)malloc(capacidad * sizeof(int));
    cola->frente = 0;
    cola->final = 0;
    cola->tamano = 0;
    cola->capacidad = capacidad;
    pthread_mutex_init(&cola->mutex, NULL);
    pthread_cond_init(&cola->no_vacia, NULL);
}

// Verifica si la cola está vacía
int esta_vacia(ColaSegura *cola)
{
    return cola->tamano == 0;
}

// Verifica si la cola está llena
int esta_llena(ColaSegura *cola)
{
    return cola->tamano == cola->capacidad;
}

// Encola un elemento
void encolar(ColaSegura *cola, int elemento)
{
    pthread_mutex_lock(&cola->mutex);

    // Espera si la cola está llena
    while (esta_llena(cola))
    {
        printf("Cola llena, productor esperando...\n");
        pthread_mutex_unlock(&cola->mutex);
        usleep(100000); // Espera 100ms
        pthread_mutex_lock(&cola->mutex);
    }

    // Agrega el elemento a la cola
    cola->elementos[cola->final] = elemento;
    cola->final = (cola->final + 1) % cola->capacidad;
    cola->tamano++;

    printf("Encolado: %d (Tamaño cola: %d)\n", elemento, cola->tamano);

    // Notifica que la cola ya no está vacía
    pthread_cond_signal(&cola->no_vacia);
    pthread_mutex_unlock(&cola->mutex);
}

// Desencola un elemento
int desencolar(ColaSegura *cola)
{
    pthread_mutex_lock(&cola->mutex);

    // Espera si la cola está vacía
    while (esta_vacia(cola))
    {
        printf("Cola vacía, consumidor esperando...\n");
        pthread_cond_wait(&cola->no_vacia, &cola->mutex);
    }

    // Elimina el elemento del frente de la cola
    int elemento = cola->elementos[cola->frente];
    cola->frente = (cola->frente + 1) % cola->capacidad;
    cola->tamano--;

    printf("Desencolado: %d (Tamaño cola: %d)\n", elemento, cola->tamano);

    pthread_mutex_unlock(&cola->mutex);
    return elemento;
}

// Libera los recursos de la cola
void destruir_cola(ColaSegura *cola)
{
    free(cola->elementos);
    pthread_mutex_destroy(&cola->mutex);
    pthread_cond_destroy(&cola->no_vacia);
}

// Función del hilo productor
void *productor(void *arg)
{
    ColaSegura *cola = (ColaSegura *)arg;

    for (int i = 1; i <= 10; i++)
    {
        int numero = rand() % 100;
        encolar(cola, numero);
        usleep(rand() % 500000); // Espera aleatoria hasta 500ms
    }

    printf("Productor finalizó\n");
    return NULL;
}

// Función del hilo consumidor
void *consumidor(void *arg)
{
    ColaSegura *cola = (ColaSegura *)arg;

    for (int i = 1; i <= 5; i++)
    {
        int elemento = desencolar(cola);
        printf("Consumidor procesó el elemento: %d\n", elemento);
        usleep(rand() % 1000000); // Espera aleatoria hasta 1s
    }

    printf("Consumidor finalizó\n");
    return NULL;
}

int main()
{
    srand(time(NULL));

    ColaSegura cola;
    inicializar_cola(&cola, TAM_MAX_COLA);

    pthread_t productores[2], consumidores[4];

    printf("Iniciando prueba de cola segura...\n");

    // Crear hilos productores
    for (int i = 0; i < 2; i++)
    {
        pthread_create(&productores[i], NULL, productor, &cola);
    }

    // Crear hilos consumidores
    for (int i = 0; i < 4; i++)
    {
        pthread_create(&consumidores[i], NULL, consumidor, &cola);
    }

    // Esperar que todos los hilos terminen
    for (int i = 0; i < 2; i++)
    {
        pthread_join(productores[i], NULL);
    }

    for (int i = 0; i < 4; i++)
    {
        pthread_join(consumidores[i], NULL);
    }

    printf("Todos los hilos finalizaron. Tamaño final de la cola: %d\n", cola.tamano);

    destruir_cola(&cola);
    return 0;
}
