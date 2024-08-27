import numpy as np

def print_npy_file_contents(file_path):
    data = np.load(file_path, allow_pickle=True)
        
    print("Data type:", type(data))
    print("Data shape:", data.shape)
    print("Data contents:")
    print(data)
    
if __name__ == "__main__":
    file_path = '0.npy'
    print_npy_file_contents(file_path)
