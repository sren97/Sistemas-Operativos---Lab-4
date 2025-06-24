#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <semaphore.h>
#include <unistd.h>
#include <time.h>

#define NUM_FILOSOFOS 5
#define TIEMPO_PENSAR 2000000 // 2 segundos
#define TIEMPO_COMER 1000000  // 1 segundo

// Estados posibles de un filósofo
typedef enum
{
    PENSANDO,
    HAMBRIENTO,
    COMIENDO
} estado_filosofos;

// Variables globales
pthread_mutex_t tenedores[NUM_FILOSOFOS];
pthread_mutex_t mutex_estado;
sem_t limite_comensales; // Límite de comensales comiendo
estado_filosofos estado[NUM_FILOSOFOS];
int comensales_comiendo = 0;

// Índice del tenedor izquierdo
int izquierdo(int id_filos)
{
    return id_filos;
}

// Índice del tenedor derecho
int derecho(int id_filos)
{
    return (id_filos + 1) % NUM_FILOSOFOS;
}

void mostrar_estado(int id, const char *accion)
{
    pthread_mutex_lock(&mutex_estado);
    printf("Filósofo %d está %s | Estados: [", id, accion);
    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        switch (estado[i])
        {
        case PENSANDO:
            printf("P");
            break;
        case HAMBRIENTO:
            printf("H");
            break;
        case COMIENDO:
            printf("C");
            break;
        }
        if (i < NUM_FILOSOFOS - 1)
            printf(" ");
    }
    printf("] Comiendo: %d\n", comensales_comiendo);
    pthread_mutex_unlock(&mutex_estado);
}

// Solución 1: Orden asimétrico para evitar interbloqueo
void tomar_tenedores_asimetrico(int id)
{
    int izq = izquierdo(id);
    int der = derecho(id);

    pthread_mutex_lock(&mutex_estado);
    estado[id] = HAMBRIENTO;
    pthread_mutex_unlock(&mutex_estado);

    mostrar_estado(id, "hambriento");

    if (id % 2 == 0)
    {
        pthread_mutex_lock(&tenedores[izq]);
        pthread_mutex_lock(&tenedores[der]);
    }
    else
    {
        pthread_mutex_lock(&tenedores[der]);
        pthread_mutex_lock(&tenedores[izq]);
    }

    pthread_mutex_lock(&mutex_estado);
    estado[id] = COMIENDO;
    comensales_comiendo++;
    pthread_mutex_unlock(&mutex_estado);

    mostrar_estado(id, "comiendo");
}

void soltar_tenedores_asimetrico(int id)
{
    int izq = izquierdo(id);
    int der = derecho(id);

    pthread_mutex_lock(&mutex_estado);
    estado[id] = PENSANDO;
    comensales_comiendo--;
    pthread_mutex_unlock(&mutex_estado);

    pthread_mutex_unlock(&tenedores[izq]);
    pthread_mutex_unlock(&tenedores[der]);

    mostrar_estado(id, "pensando");
}

// Solución 2: Semáforo para limitar comensales
void tomar_tenedores_semaforo(int id)
{
    pthread_mutex_lock(&mutex_estado);
    estado[id] = HAMBRIENTO;
    pthread_mutex_unlock(&mutex_estado);

    mostrar_estado(id, "hambriento");

    sem_wait(&limite_comensales);

    int izq = izquierdo(id);
    int der = derecho(id);

    if (izq < der)
    {
        pthread_mutex_lock(&tenedores[izq]);
        pthread_mutex_lock(&tenedores[der]);
    }
    else
    {
        pthread_mutex_lock(&tenedores[der]);
        pthread_mutex_lock(&tenedores[izq]);
    }

    pthread_mutex_lock(&mutex_estado);
    estado[id] = COMIENDO;
    comensales_comiendo++;
    pthread_mutex_unlock(&mutex_estado);

    mostrar_estado(id, "comiendo");
}

void soltar_tenedores_semaforo(int id)
{
    int izq = izquierdo(id);
    int der = derecho(id);

    pthread_mutex_lock(&mutex_estado);
    estado[id] = PENSANDO;
    comensales_comiendo--;
    pthread_mutex_unlock(&mutex_estado);

    pthread_mutex_unlock(&tenedores[izq]);
    pthread_mutex_unlock(&tenedores[der]);

    sem_post(&limite_comensales);
    mostrar_estado(id, "pensando");
}

void *filosofo_asimetrico(void *arg)
{
    int id = *((int *)arg);
    for (int i = 0; i < 3; i++)
    {
        mostrar_estado(id, "pensando");
        usleep(rand() % TIEMPO_PENSAR);
        tomar_tenedores_asimetrico(id);
        usleep(rand() % TIEMPO_COMER);
        soltar_tenedores_asimetrico(id);
    }
    printf("Filósofo %d terminó de comer\n", id);
    return NULL;
}

void *filosofo_semaforo(void *arg)
{
    int id = *((int *)arg);
    for (int i = 0; i < 3; i++)
    {
        mostrar_estado(id, "pensando");
        usleep(rand() % TIEMPO_PENSAR);
        tomar_tenedores_semaforo(id);
        usleep(rand() % TIEMPO_COMER);
        soltar_tenedores_semaforo(id);
    }
    printf("Filósofo %d terminó de comer\n", id);
    return NULL;
}

void ejecutar_asimetrico()
{
    printf("\n=== Solución con Orden Asimétrico ===\n");
    pthread_t hilos[NUM_FILOSOFOS];
    int ids[NUM_FILOSOFOS];

    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        estado[i] = PENSANDO;
        ids[i] = i;
        pthread_create(&hilos[i], NULL, filosofo_asimetrico, &ids[i]);
    }
    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        pthread_join(hilos[i], NULL);
    }
    printf("Solución asimétrica finalizada\n");
}

void ejecutar_semaforo()
{
    printf("\n=== Solución con Semáforo ===\n");
    pthread_t hilos[NUM_FILOSOFOS];
    int ids[NUM_FILOSOFOS];

    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        estado[i] = PENSANDO;
        ids[i] = i;
    }
    comensales_comiendo = 0;

    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        pthread_create(&hilos[i], NULL, filosofo_semaforo, &ids[i]);
    }
    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        pthread_join(hilos[i], NULL);
    }
    printf("Solución con semáforo finalizada\n");
}

int main()
{
    srand(time(NULL));

    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        pthread_mutex_init(&tenedores[i], NULL);
    }
    pthread_mutex_init(&mutex_estado, NULL);
    sem_init(&limite_comensales, 0, NUM_FILOSOFOS - 1);

    printf("Problema de los Filósofos Comensales\n");
    printf("Filósofos: %d | Cada uno comerá 3 veces\n", NUM_FILOSOFOS);

    ejecutar_asimetrico();
    sleep(2);
    ejecutar_semaforo();

    for (int i = 0; i < NUM_FILOSOFOS; i++)
    {
        pthread_mutex_destroy(&tenedores[i]);
    }
    pthread_mutex_destroy(&mutex_estado);
    sem_destroy(&limite_comensales);

    return 0;
}
