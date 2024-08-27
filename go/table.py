import pandas as pd
import numpy as np

np.random.seed(42)

df = pd.DataFrame(np.random.randint(0, 2**3 + 1, size=(8, 8)))

csv_file_path = "lookupTables/MFIP-Rand/MFIPSubRandPos_nB_3_dimF_512.csv"
df.to_csv(csv_file_path, index=False)
