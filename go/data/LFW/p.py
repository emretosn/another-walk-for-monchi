import numpy as np
import csv
import os


directory = './John_Lennon'


for filename in os.listdir(directory):

    if filename.endswith('.npy'):
        npy_file_path = os.path.join(directory, filename)


        data = np.load(npy_file_path)


        csv_file_path = os.path.join(directory, filename.replace('.npy', '.csv'))


        with open(csv_file_path, 'w', newline='') as csvfile:
            writer = csv.writer(csvfile)


            if data.ndim == 1:
                for item in data:
                    writer.writerow([item])


            else:
                for row in data:
                    writer.writerow(row)

        print(f"Converted {filename} to {csv_file_path}")

print("All .npy files have been successfully converted to .csv files.")

