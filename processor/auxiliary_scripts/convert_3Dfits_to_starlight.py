#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Mon Apr  8 11:09:17 2024

@author: mario
"""

import numpy as np
from astropy.io import fits
import argparse


# Parser
parser = argparse.ArgumentParser(description='Converts 3D FITS files to a format compatible with Starlight')
parser.add_argument('-s', '--spectrum', metavar='3D DATACUBE', help='3D datacube', type=argparse.FileType('rb'))
parser.add_argument('-o', '--output-directory', metavar='STRING ID', default='./', help='Directory to save the output files')
parser.add_argument('-i', '--output-file', metavar='STRING ID', default='test', help='Label added to saved output files')


args = parser.parse_args()


# meg_spectra = "/Users/mario/AC3/scripts_auxiliares/prueba/NGC6027_LR-V_final_cube.fits"

# Path to the RSS spectrum file
meg_spectra = args.spectrum.name

# Open the RSS spectrum FITS file
hdu = fits.open(meg_spectra)

# Extracting data and header
gal_lin = hdu[0].data   ### Flux en Jy ###
h1_0 = hdu[0].header

# Define wavelength range
lam_gal = h1_0['CRVAL3'] + h1_0['CDELT3']*np.arange(h1_0['NAXIS3'])
lamRange1 = h1_0['CRVAL3'] + np.array([0., h1_0['CDELT3']*(h1_0['NAXIS3'] - 1)])

# Dimensions of the RSS spectrum
x_num = gal_lin.shape[1]
y_num = gal_lin.shape[2]
z_num = gal_lin.shape[0]
rows_number = x_num*y_num
cifras_x_num = len(str(x_num))
cifras_y_num = len(str(y_num))
cifras_rows_number = len(str(rows_number))
columns_number = z_num

# Column names for the output file
column_names = "lambda flux"

# Output directory and file
output_directory = args.output_directory
output_file = args.output_file

# Loop through all spaxels
counter = 0
for i in range(x_num):
    for j in range(y_num):
        counter += 1
        print(f"Spectra converted: {counter}/{rows_number}", end='\r')
        spaxel_spectrum = gal_lin[:,i,j]
        spaxel_spectrum_cgs = spaxel_spectrum*1e17 / (3.33564095e4*(lam_gal**2))   ## Units 10**-17 erg cm-2 s-1 Angstrom-1
        spaxel_spectrum_saved = np.maximum(spaxel_spectrum_cgs, 0)
        # Save each spectrum to a text file
        np.savetxt((output_directory + 'spectrum_xpos_' + str(i).zfill(cifras_x_num) + '_ypos_' + str(j).zfill(cifras_y_num) + "_" + output_file + '.txt'), np.column_stack((lam_gal, spaxel_spectrum_saved)), header = 'lambda flux', comments='', delimiter = ' ', fmt = '%.1f %.14f')
