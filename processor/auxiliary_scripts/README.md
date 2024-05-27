# Auxiliary Scripts

This folder contains two auxiliary scripts that provide additional functionality for the main project.

## Script 1: convert_3Dfits_to_starlight.py

This script converts 3D FITS files to a format compatible with Starlight.

### Usage

```

python convert_3Dfits_to_starlight.py [options] [-h] [-s 3D DATACUBE] [-o OUTPUT DIRECTORY] [-i OUTPUT FILE]

Arguments:
-h, --help: Show help message and exit.
-s 3D DATACUBE, --spectrum 3D DATACUBE: Path to the 3D datacube.
-o OUTPUT DIRECTORY, --output-directory OUTPUT DIRECTORY: Directory to save the output files.
-i OUTPUT FILE, --output-file OUTPUT FILE: Label added to saved output files.

Example:

python convert_3Dfits_to_starlight.py -s datacube.fits -o output_dir -i output_file
```

## Script 2: starlight_grid_file_assign_spectra.py

This script modifies the `grid.in` file used by Starlight for analysis by automatically including the spectra you want to analyze.

## Usage

```
starlight_grid_file_assign_spectra.py [-h] input_file output_file [spectrum_names [spectrum_names ...]]

Arguments:
input_file: Path to the input file (grid.in).
output_file: Path to the output file (modified grid.in).
spectrum_names: Names of spectrum files to include in the analysis.
-h, --help: Show help message and exit.

Example:

python starlight_grid_file_assign_spectra.py grid.in modified_grid.in spectrum1.txt spectrum2.txt spectrum3.txt

python starlight_grid_file_assign_spectra.py grid.in modified_grid.in spectra_folder/*