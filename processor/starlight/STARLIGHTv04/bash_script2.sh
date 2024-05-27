#!/bin/bash
set -x

# Ruta al ejecutable y al archivo de entrada
EXECUTABLE="/docker/starlight/STARLIGHTv04/StarlightChains_v04.amd64_g77-3.4.6-r1_static.exe"
INPUT_FILE="/starlight/config_files_starlight/grid_example.in"
DATA_FILE_FLAG="/starlight/start_starlight"


# Verificar si el ejecutable existe
if [ ! -f "$EXECUTABLE" ]; then
    echo "Error: No se encontró el ejecutable $EXECUTABLE"
    exit 1
fi

echo "Running Executable"
# Ejecutar el ejecutable con el archivo de entrada
#"$EXECUTABLE" < "$INPUT_FILE"


while :
do
    #if flag set
    if [ ! -f "$DATA_FILE_FLAG" ]; then
        echo "Waiting for data file to start"
    else
        echo "Starting Application"
        ./StarlightChains_v04.amd64_g77-3.4.6-r1_static.exe < /starlight/grid_example.in
        exit_code=$?

        if [ $exit_code -ne 0 ]; then
            echo "Error"
        fi
        
        echo "Removing Start Flag"
        rm "$DATA_FILE_FLAG"
        echo "Complete"
        ls -al "$DATA_FILE_FLAG"
    fi
    sleep 10
done

## need to handle errors, remove flag or flag error