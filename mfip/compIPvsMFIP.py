from itertools import repeat
import multiprocessing as mp
from datetime import datetime
import numpy as np
import pickle
import gc
import os

from utils import *
from mfbrTabs import *


def main():
    now = datetime.now()
    timeTag = now.strftime("%d%m%Y_%H%M%S")

    bordersDir = f'./lookupTables/Borders/'
    tabQMFIPdir = f'./lookupTables/MFIPSubRand/'
    tabRanddir = f'./lookupTables/Rand/'

    numSamples = 2*100000
    idSynSamples = np.arange(0,numSamples,2, dtype=int)

    print('Normalized samples\nnumSamples = ', numSamples)

    dimFlist = [32, 64, 128]
    nBList = np.arange(2,4)
    dQList = [0.001]

    for dimF in dimFlist:
        pearson = {d:[] for d in dQList}

        resDir = f'./results/Synthetic/IP/dimF_{dimF}'
        os.makedirs(resDir, exist_ok=True)

        (synSamples) = pickle.load(open(f'./data/Synthetic/syntheticSamples_dimF_{dimF}.pkl', 'rb'))

        print(f'IP Synthetic Samples dimF = {dimF}')

        for nB in nBList:
            borders = pickle.load(open(f'{bordersDir}/Borders_nB_{nB}_dimF_{dimF}.pkl', 'rb'))
            tabQMFIP = pickle.load(open(f'{tabQMFIPdir}/MFIPSubRand_nB_{nB}_dimF_{dimF}.pkl', 'rb'))
            tabRand = pickle.load(open(f'{tabRanddir}/Rand_nB_{nB}_dimF_{dimF}.pkl', 'rb'))

            for dQ in dQList:
                fct_args = zip(idSynSamples, repeat(synSamples), repeat(borders), repeat(tabQMFIP), repeat(tabRand))

                pool = mp.Pool(32)
                scoresIPandIPQ = pool.starmap(compIPandIPQ, fct_args)
                scoresIP, scoresIPQ = zip(*scoresIPandIPQ)
                gc.collect()

                scoresIP = np.array(scoresIP, dtype = float)
                scoresIPQ = np.array(scoresIPQ, dtype = int)

                pool.close()
                pool.join()

                r = stats.pearsonr(scoresIPQ, scoresIP)[0]
                pearson[dQ].append(r)

                r = np.round(r, 4)
                plotTitle = f'dimF = {dimF} levels = 2^{nB} dQ = {dQ} Pearson = {r}'
                print(plotTitle)
                plotFile = f'{resDir}/Scatter_IPvsMFIP_dimF_{dimF}_nB_{nB}_dQ_{dQ}_{timeTag}.pdf'
                plotScatter(plotFile, plotTitle, scoresIP, scoresIPQ, 'IP', 'IP Quantized')

        plotPearson(f'{resDir}/Pearson_IP_{dimF}.pdf', f'MFIP (dimF = {dimF})', nBList, pearson)
        pickle.dump((nBList, dQList, pearson), open(f'{resDir}/results_dimF_{dimF}_{timeTag}.pkl', 'wb'))


if __name__ == '__main__':
    main()



























