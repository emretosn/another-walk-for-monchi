import os
import numpy as np
import csv

# Define the main folder path
main_folder_path = '.'

for root, dirs, files in os.walk(main_folder_path):
    for file in files:
        if file.endswith('.npy'):
            # Load the .npy file
            npy_file_path = os.path.join(root, file)
            data = np.load(npy_file_path)

            # Ensure the data is flattened (1D array)
            flat_data = data.flatten()

            # Save to CSV file
            csv_file_path = os.path.join(root, file.replace('.npy', '.csv'))
            with open(csv_file_path, 'w', newline='') as csv_file:
                writer = csv.writer(csv_file)

                # Write each element as separate entries in the same row
                writer.writerow(flat_data)

            print(f'Saved {csv_file_path}')

