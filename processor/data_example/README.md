# Starlight Data Preparation

This folder contains data files necessary for running Starlight analysis.

## NGC6027_LR-V_final_cube.fits

This file (`NGC6027_LR-V_final_cube.fits`) is an observation of the galaxy NGC6027 in a 3D datacube format stored in a FITS file. It represents the spatial and spectral information of the galaxy.

## starlight_format

This subfolder contains individual spectra extracted from the data cube (`NGC6027_LR-V_final_cube.fits`). These spectra are stored in text files (`*.txt`) and are formatted for use with Starlight software. Each text file contains the spectrum corresponding to a specific position (x, y) in the data cube, and its filename indicates the coordinates (xpos, ypos) from which the spectrum was extracted.
