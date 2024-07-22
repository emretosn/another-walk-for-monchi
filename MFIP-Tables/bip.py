from itertools import repeat, product
import multiprocessing as mp
import numpy as np
import pickle
import gc
import os

RANDN = 16
dQ = 0.001
def map_range(value, original_min, original_max, target_min, target_max):
    # Map the value from the original range to the target range
    return ((value - original_min) / (original_max - original_min)) * (target_max - target_min) + target_min

def genStoreRandMFIPSubstRand(nBdimF, tabMFIPdir, tabRanddir, tabMFIPSubRanddir):
    nB, dimF = nBdimF

    tabShare1 = np.random.randint(0, 2**RANDN, (2**nB, 2**nB), dtype= np.uint32)
    pickle.dump((tabShare1), open(f'{tabRanddir}/Rand_nB_{nB}_dimF_{dimF}.pkl', 'wb'))

    with open(f'{tabMFIPdir}/MFIP_nB_{nB}_dimF_{dimF}.pkl', 'rb') as f:
        tabMFIP = pickle.load(f)
    tabQMFIP = np.round(tabMFIP/dQ).astype(int)
    tabQMFIP = map_range(tabQMFIP, np.min(tabQMFIP), np.max(tabQMFIP), 0, 2**RANDN).astype(np.uint32)
    with open(f'{tabRanddir}/Rand_nB_{nB}_dimF_{dimF}.pkl', 'rb') as f:
        tabShare1 = pickle.load(f)
    # max number of bits tabShare1 
    tabShare2 = tabQMFIP - tabShare1
    assert np.equal(tabQMFIP, tabShare1 + tabShare2).all()
    pickle.dump(tabShare2, open(f'{tabMFIPSubRanddir}/MFIPSubRand_nB_{nB}_dimF_{dimF}.pkl', 'wb'))

def main():
    nBList = np.arange(2,4)
    dimFlist = [32, 64, 128, 256, 512]
    nBdimF = product(nBList, dimFlist)

    tabMFIPdir = f'./lookupTables/MFIP/'
    tabRanddir = f'./lookupTables/Rand/'
    tabMFIPSubRanddir = f'./lookupTables/MFIPSubRand/'

    os.makedirs(tabRanddir, exist_ok=True)
    os.makedirs(tabMFIPdir, exist_ok=True)
    os.makedirs(tabMFIPSubRanddir, exist_ok=True)

    fct_args = zip(nBdimF, repeat(tabMFIPdir), repeat(tabRanddir), repeat(tabMFIPSubRanddir))

    print('Start pool.starmap')
    pool = mp.Pool(32)

    pool.starmap(genStoreRandMFIPSubstRand, fct_args)
    gc.collect()

    print('End pool.starmap')

    pool.close()
    pool.join()


if __name__ == '__main__':
    main()


