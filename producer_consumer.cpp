#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <semaphore.h>
#include <unistd.h>
#include <time.h>

#define TAM_BUFFER 10
#define NUM_PRODUCTORES 2
#define NUM_CONSUMIDORES 3
#define ELEMENTOS_POR_PRODUCTOR 8

int buffer[TAM_BUFFER];
int entrada = 0; // Índice para insertar
int salida = 0;  // Índice para consumir

sem_t vacios;                 // Contador de espacios vacíos
sem_t llenos;                 // Contador de espacios llenos
pthread_mutex_t mutex_buffer; // Protege acceso al buffer

// Contadores globales
int total_producidos = 0;
int total_consumidos = 0;
pthread_mutex_t mutex_contador;

// Genera un nuevo elemento (simulado)
int producir_elemento()
{
    return rand() % 1000;
}

// Procesa un elemento (simulado)
void consumir_elemento(int elemento)
{
    printf("Consumido: %d\n", elemento);
    usleep(rand() % 500000); // Espera simulada
}

// Función del hilo productor
void *productor(void *arg)
{
    int id = *((int *)arg);

    for (int i = 0; i < ELEMENTOS_POR_PRODUCTOR; i++)
    {
        int elemento = producir_elemento();

        sem_wait(&vacios); // Espera espacio disponible

        pthread_mutex_lock(&mutex_buffer);

        buffer[entrada] = elemento;
        printf("Productor %d produjo %d en posición %d\n", id, elemento, entrada);
        entrada = (entrada + 1) % TAM_BUFFER;

        pthread_mutex_lock(&mutex_contador);
        total_producidos++;
        pthread_mutex_unlock(&mutex_contador);

        pthread_mutex_unlock(&mutex_buffer);

        sem_post(&llenos); // Señala espacio lleno

        usleep(rand() % 300000); // Simula tiempo de producción
    }

    printf("Productor %d finalizó\n", id);
    return NULL;
}

// Función del hilo consumidor
void *consumidor(void *arg)
{
    int id = *((int *)arg);
    int consumidos_local = 0;

    while (1)
    {
        sem_wait(&llenos); // Espera si no hay elementos

        pthread_mutex_lock(&mutex_buffer);

        pthread_mutex_lock(&mutex_contador);
        if (total_consumidos >= NUM_PRODUCTORES * ELEMENTOS_POR_PRODUCTOR)
        {
            pthread_mutex_unlock(&mutex_contador);
            pthread_mutex_unlock(&mutex_buffer);
            break;
        }

        int elemento = buffer[salida];
        printf("Consumidor %d consumió %d de posición %d\n", id, elemento, salida);
        salida = (salida + 1) % TAM_BUFFER;

        total_consumidos++;
        consumidos_local++;
        pthread_mutex_unlock(&mutex_contador);

        pthread_mutex_unlock(&mutex_buffer);

        sem_post(&vacios); // Libera un espacio

        consumir_elemento(elemento);
    }

    printf("Consumidor %d finalizó (consumió %d elementos)\n", id, consumidos_local);
    return NULL;
}

// Muestra el estado del buffer
void mostrar_estado_buffer()
{
    int espacios_vacios, espacios_llenos;
    sem_getvalue(&vacios, &espacios_vacios);
    sem_getvalue(&llenos, &espacios_llenos);

    printf("\n=== Estado del Buffer ===\n");
    printf("Espacios vacíos: %d, llenos: %d\n", espacios_vacios, espacios_llenos);
    printf("Total producidos: %d, consumidos: %d\n", total_producidos, total_consumidos);
    printf("Contenido: [");
    for (int i = 0; i < TAM_BUFFER; i++)
    {
        if (i == entrada)
            printf("ENT->");
        if (i == salida)
            printf("SAL->");
        printf("%d", buffer[i]);
        if (i < TAM_BUFFER - 1)
            printf(", ");
    }
    printf("]\n");
    printf("=========================\n\n");
}

int main()
{
    srand(time(NULL));

    sem_init(&vacios, 0, TAM_BUFFER); // Todos vacíos inicialmente
    sem_init(&llenos, 0, 0);          // Ninguno lleno inicialmente

    pthread_mutex_init(&mutex_buffer, NULL);
    pthread_mutex_init(&mutex_contador, NULL);

    // Inicializa buffer
    for (int i = 0; i < TAM_BUFFER; i++)
    {
        buffer[i] = 0;
    }

    printf("Simulación Productores-Consumidores\n");
    printf("Tamaño del buffer: %d\n", TAM_BUFFER);
    printf("Productores: %d (cada uno produce %d)\n", NUM_PRODUCTORES, ELEMENTOS_POR_PRODUCTOR);
    printf("Consumidores: %d\n", NUM_CONSUMIDORES);
    printf("Total a producir: %d elementos\n\n", NUM_PRODUCTORES * ELEMENTOS_POR_PRODUCTOR);

    pthread_t productores[NUM_PRODUCTORES];
    pthread_t consumidores[NUM_CONSUMIDORES];
    int ids_productores[NUM_PRODUCTORES];
    int ids_consumidores[NUM_CONSUMIDORES];

    // Crear hilos productores
    for (int i = 0; i < NUM_PRODUCTORES; i++)
    {
        ids_productores[i] = i + 1;
        pthread_create(&productores[i], NULL, productor, &ids_productores[i]);
    }

    // Crear hilos consumidores
    for (int i = 0; i < NUM_CONSUMIDORES; i++)
    {
        ids_consumidores[i] = i + 1;
        pthread_create(&consumidores[i], NULL, consumidor, &ids_consumidores[i]);
    }

    // Esperar a que los productores terminen
    for (int i = 0; i < NUM_PRODUCTORES; i++)
    {
        pthread_join(productores[i], NULL);
    }

    printf("\nTodos los productores terminaron. Esperando consumidores...\n");

    // Libera los semáforos llenos para que los consumidores puedan continuar
    for (int i = 0; i < NUM_CONSUMIDORES; i++)
    {
        sem_post(&llenos); 
    }

    // Esperar a que los consumidores terminen
    for (int i = 0; i < NUM_CONSUMIDORES; i++)
    {
        pthread_join(consumidores[i], NULL);
    }

    printf("\nTodos los hilos terminaron.\n");
    mostrar_estado_buffer();

    // Liberar recursos
    sem_destroy(&vacios);
    sem_destroy(&llenos);
    pthread_mutex_destroy(&mutex_buffer);
    pthread_mutex_destroy(&mutex_contador);

    return 0;
}
