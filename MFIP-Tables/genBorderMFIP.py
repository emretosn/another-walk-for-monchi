from itertools import repeat, product
import multiprocessing as mp
import numpy as np
import pickle
import gc
import os

from utils import *
from mfbrTabs import *

'''
         REMOVED ALL MFSED
'''

#def genStoreBordersMFIPandMFSED(nBDimF, bordersDir, tabMFIPdir, tabMFSEDdir):
def genStoreBordersMFIPandMFSED(nBDimF, bordersDir, tabMFIPdir):
    nB, dimF = nBDimF
    #borders, tabMFIP, tabMFSED = genBordersLookupTables(nB, dimF)
    borders, tabMFIP = genBordersLookupTables(nB, dimF)
    pickle.dump((borders), open(f'{bordersDir}/Borders_nB_{nB}_dimF_{dimF}.pkl', 'wb'))
    pickle.dump((tabMFIP), open(f'{tabMFIPdir}/MFIP_nB_{nB}_dimF_{dimF}.pkl', 'wb'))
    #pickle.dump((tabMFSED), open(f'{tabMFSEDdir}/MFSED_nB_{nB}_dimF_{dimF}.pkl', 'wb'))
    print(f'Borders and MFIP tables for dimension = {dimF} and feature levels 2^{nB} are generated and saved.')


def main():
    dimFlist = [32, 64, 128, 256, 512]
    # Changed from (2,13) to (2,4) for the sake of the demo
    #nBList = np.arange(2,13)
    #nBList = np.arange(2,4)
    nBList = [16]
    nBDimF = product(nBList, dimFlist)

    bordersDir = f'./lookupTables/Borders/'
    tabMFIPdir = f'./lookupTables/MFIP/'
    #tabMFSEDdir = f'./lookupTables/MFSED/'

    os.makedirs(bordersDir, exist_ok=True)
    os.makedirs(tabMFIPdir, exist_ok=True)
    #os.makedirs(tabMFSEDdir, exist_ok=True)

    #fct_args = zip(nBDimF, repeat(bordersDir), repeat(tabMFIPdir), repeat(tabMFSEDdir))
    fct_args = zip(nBDimF, repeat(bordersDir), repeat(tabMFIPdir))

    print('Start pool.starmap')
    pool = mp.Pool(32)

    pool.starmap(genStoreBordersMFIPandMFSED, fct_args)
    gc.collect()

    print('End pool.starmap')

    pool.close()
    pool.join()


if __name__ == '__main__':
    main()



