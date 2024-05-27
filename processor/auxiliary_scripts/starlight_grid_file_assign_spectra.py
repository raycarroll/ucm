import os
import argparse

def modify_and_remove_lines_below(input_file, output_file, spectrum_names):
    lines_to_keep = []
    flag_found = False
    with open(input_file, 'r') as f_input:
        for line in f_input:
            if flag_found:
                break
            if "IsFlagSpecAvailable" in line:
                flag_found = True
            lines_to_keep.append(line)

    with open(output_file, 'w') as f_output:
        for line in lines_to_keep:
            f_output.write(line)

        # Añadir nuevas líneas al final del archivo
        for spectrum_name_with_ext in spectrum_names:
            spectrum_name = os.path.splitext(os.path.basename(spectrum_name_with_ext))[0]  # Obtener solo el nombre del archivo sin la extensión ni la ruta
            new_line = f"{spectrum_name}.txt StCv04.C11.arp220.config Base.BC03.N Masks.Em.Abs.Lines.Arp220.gm CAL 37.755 0.555 {spectrum_name}_output.txt\n"
            f_output.write(new_line)

def parse_arguments():
    parser = argparse.ArgumentParser(description="Modify a file and add new lines at the end")
    parser.add_argument("input_file", help="Path to the input file")
    parser.add_argument("output_file", help="Path to the output file")
    parser.add_argument("spectrum_names", nargs="*", help="Names of spectrum files")
    return parser.parse_args()

if __name__ == "__main__":
    args = parse_arguments()
    modify_and_remove_lines_below(args.input_file, args.output_file, args.spectrum_names)



