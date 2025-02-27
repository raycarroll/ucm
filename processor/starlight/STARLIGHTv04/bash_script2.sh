#!/bin/bash


set -x

# Path to the executable and process list
EXECUTABLE="/docker/starlight/STARLIGHTv04/StarlightChains_v04.amd64_g77-3.4.6-r1_static.exe"
PROCESS_FILE="/starlight/runtime/processlist.txt"
INFILES_DIR="/starlight/runtime/infiles"

# Function to remove the first line from the process list
removeInFileFromList() {
    echo "Before removal:"
    cat "$PROCESS_FILE"
    sed -i '1d' "$PROCESS_FILE"
    echo "After removal:"
    cat "$PROCESS_FILE"
}

# Function to check if a file exists and is readable
file_exists_and_readable() {
    if [[ -f "$1" && -r "$1" ]]; then
        return 0
    else
        return 1
    fi
}


while :
do
    echo "Reading next line from process list..."
    firstline=$(head -n 1 "$PROCESS_FILE" | xargs) 

    if [[ -z "$firstline" ]]; then
        echo "Process list is empty. Waiting for new files..."
    else
        echo "Next file to process: $firstline"

        # Check if the .in file exists
        in_file_path="$INFILES_DIR/$firstline"
        if file_exists_and_readable "$in_file_path"; then
            echo "Starting application with input file: $in_file_path"

            # Run the executable with the .in file as input
            "$EXECUTABLE" < "$in_file_path"
            exit_code=$?

            if [ $exit_code -ne 0 ]; then
                echo "Error: Application failed with exit code $exit_code"
            else
                echo "Application completed successfully"
            fi

            # Remove the processed .in file from the process list
            echo "Removing processed file from the process list..."
            removeInFileFromList
        else
            echo "Error: Input file $in_file_path does not exist or is not readable"
            # Remove the invalid entry from the process list to avoid blocking the queue
            removeInFileFromList
        fi
    fi

    # Sleep for a while before checking the process list again
    sleep 10
done