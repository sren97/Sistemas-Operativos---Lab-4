
Practice # 4- Synchronization Mechanisms in Operating Systems
Sebastián Rentería Palacios
David Agudelo Ochoa

Este proyecto implementa tres problemas fundamentales de sincronización en sistemas operativos usando diferentes primitivas de sincronización.
Descripción General
La implementación incluye:

Cola Thread-Safe usando mutexes y variables de condición
Problema Productor-Consumidor usando semáforos
Problema de los Filósofos Comensales con prevención de deadlock


Prerrequisitos

Compilador GCC con soporte pthread
Entorno Linux/Unix o WSL
Soporte para semáforos POSIX

Para compilar archivos .cpp use:
g++ -o nombre_ejecutable.exe archivo.cpp

Para compilar archivos go, use:
go build archivo.go


Compilación y Ejecución


ejecutar individualmente
./queue
./producer_consumer
./dining_philosophers