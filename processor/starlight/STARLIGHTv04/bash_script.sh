#!/bin/bash

# Ruta al ejecutable y al archivo de entrada
EXECUTABLE="/home/starlight/STARLIGHTv04/StarlightChains_v04.amd64_g77-3.4.6-r1_static.exe"
INPUT_FILE="/home/starlight/shared_directory/config_files_starlight/grid_example.in"


# Verificar si el ejecutable existe
if [ ! -f "$EXECUTABLE" ]; then
    echo "Error: No se encontró el ejecutable $EXECUTABLE"
    exit 1
fi

# Verificar si el archivo de entrada existe
if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: No se encontró el archivo de entrada $INPUT_FILE"
    exit 1
fi

echo "Running Executable"
# Ejecutar el ejecutable con el archivo de entrada
#"$EXECUTABLE" < "$INPUT_FILE"

 ./StarlightChains_v04.amd64_g77-3.4.6-r1_static.exe < /home/starlight/shared_directory/config_files_starlight/grid_example.in