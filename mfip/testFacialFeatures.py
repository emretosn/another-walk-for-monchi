import pickle
import gc

import numpy as np
from itertools import repeat
import multiprocessing as mp

from datetime import datetime

import os

from mfbrTabs import *
from utils import *


def main():
    now = datetime.now()
    timeTag = now.strftime("%d%m%Y_%H%M%S")

    dataset = 'VGGFace2'
    testType = 'test'
    modelType = 'ArcFace-R100' # 'CosFace-R100' # 'ArcFace-R100'
    similarity = 'IP'

    dimF = 512
    nB = 3 # 4 6
    dQ = 0.001 # 0.001 # 0.01
    precisionD = 1E-6

    dataDir = f'./data/{dataset}/{testType}/{modelType}'
    resDir = f'./results/{dataset}/{modelType}/nB_{nB}'
    os.makedirs(resDir, exist_ok=True)

    subjectIDs = os.listdir(dataDir)
    subjectIDs.sort()

    print(f'Baseline {similarity} vs MF{similarity} tested on {dataset} dataset extracted with {modelType}')

    bordersFile = f'./lookupTables/Borders/Borders_nB_{nB}_dimF_{dimF}.pkl'
    tabQMFIP = f'lookupTables/MFIPSubRand/MFIPSubRand_nB_{nB}_dimF_{dimF}.pkl'
    tabRand = f'./lookupTables/Rand/Rand_nB_{nB}_dimF_{dimF}.pkl'

    borders = pickle.load(open(bordersFile, "rb"))
    tabQMF = pickle.load(open(tabQMFIP, "rb"))
    tabRand = pickle.load(open(tabRand, "rb"))

    print('-- Run mated and non-mated comparison for ', similarity)

    pool = mp.Pool(32)

    gen_args = zip(repeat(dataDir), subjectIDs, repeat(borders), repeat(tabQMF), repeat(tabRand))

    imp_args = zip(repeat(dataDir), subjectIDs, repeat(subjectIDs), repeat(borders), repeat(tabQMF), repeat(tabRand))

    print('Start pool.starmap')

    print(f'{similarity} mated')
    matedScBaseandMF = pool.starmap(matedIP, gen_args)
    gc.collect()

    print(f'{similarity} non-mated')
    nonMatedSCBaseandMF = pool.starmap(nonMatedIP, imp_args)
    gc.collect()

    print('End pool.starmap')

    matedScBase, matedScMFQ = zip(*matedScBaseandMF)
    matedScBase = flattenList(matedScBase)
    matedScMFQ = flattenList(matedScMFQ)
    matedScBase = np.array(matedScBase, dtype = float)
    matedScMFQ = np.array(matedScMFQ, dtype = int)
    gc.collect()

    nonMatedScBase, nonMatedScMFQ = zip(*nonMatedSCBaseandMF)
    nonMatedScBase = flattenList(nonMatedScBase)
    nonMatedScMFQ = flattenList(nonMatedScMFQ)
    nonMatedScBase = np.array(nonMatedScBase, dtype = float)
    nonMatedScMFQ = np.array(nonMatedScMFQ, dtype = int)
    gc.collect()

    pool.close()
    pool.join()


    print(f'Baseline {similarity}: #Mated_Comparisons = {len(matedScBase)} and #NonMated_Comparisons = {len(nonMatedScBase)}')
    print(f'Baseline {similarity}: Mated score range = [{min(matedScBase)}, {max(matedScBase)}]')
    print(f'Baseline {similarity}: NonMated score range = [{min(nonMatedScBase)}, {max(nonMatedScBase)}]')


    print(f'MF{similarity}: #Mated_Comparisons = {len(matedScMFQ)} and #NonMated_Comparisons = {len(nonMatedScMFQ)}')
    print(f'MF{similarity}: Mated score range = [{min(matedScMFQ)}, {max(matedScMFQ)}]')
    print(f'MF{similarity}: NonMated score range = [{min(nonMatedScMFQ)}, {max(nonMatedScMFQ)}]')


    print('-- Measure the performance')


    _, fmrBase, tmrBase = get_fmr_tmr_Prec(matedScBase, nonMatedScBase, precision = precisionD)
    eerBase = calc_eerPT(fmrBase, tmrBase)
    print(f'{similarity} Base eer = ', eerBase)
    gc.collect()

    _, fmrMFQ, tmrMFQ = get_fmr_tmr_Prec(matedScMFQ, nonMatedScMFQ, precision = 1)
    eerMFQ = calc_eerPT(fmrMFQ, tmrMFQ)
    print(f'MF{similarity} eer = ', eerMFQ)
    gc.collect()


    resultsFile = resDir+f'/Results_{dataset}_{modelType}_{similarity}_nB_{nB}_dQ_{dQ}_{timeTag}.pkl'
    pickle.dump((matedScBase, nonMatedScBase, fmrBase, tmrBase, eerBase,
    matedScMFQ, nonMatedScMFQ, fmrMFQ, tmrMFQ, eerMFQ), open(resultsFile, 'wb'))

    plotFile = resDir+f'/DET_{dataset}_{modelType}_{similarity}_nB_{nB}_dQ_{dQ}_{timeTag}'
    plotTitle = f'{similarity} DET curve of {dataset} {modelType} dataset'

    print(plotFile)

    fmrList = [fmrBase, fmrMFQ]
    tmrList = [tmrBase, tmrMFQ]
    labelList = [f'{similarity} (EER = {eerBase:.3E})', f'MF{similarity} (EER = {eerMFQ:.3E})']
    linewidthList = [2, 2]

    plot_MultipleDETsSaved(plotFile+'.pdf', plotTitle, fmrList, tmrList, labelList, linewidthList)


if __name__ == '__main__':
    main()

