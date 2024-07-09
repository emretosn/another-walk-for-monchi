

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
    tabMFSEDdir = f'./lookupTables/MFSED/'  
    
    
    
    numSamples = 2*100000 
    idSynSamples = np.arange(0,numSamples,2, dtype=int)

    print('Normalized samples\nnumSamples = ', numSamples)
    
    dimFlist = [32, 64, 128, 256, 512]
    nBList = np.arange(2,13) 
    dQList = [0.01, 0.001] # [0.1, 0.01, 0.001, 0.0001, 0.00001, 0.000001]

        
    for dimF in dimFlist:   
        
        pearson = {d:[] for d in dQList}        
        
        resDir = f'./results/Synthetic/SED/dimF_{dimF}'
        os.makedirs(resDir, exist_ok=True)
        
        (synSamples) = pickle.load(open(f'./data/Synthetic/syntheticSamples_dimF_{dimF}.pkl', 'rb'))        
        
        print(f'SED Synthetic Samples dimF = {dimF}')         
        
        for nB in nBList:            
            borders = pickle.load(open(f'{bordersDir}/Borders_nB_{nB}_dimF_{dimF}.pkl', 'rb')) 
            tabMFSED = pickle.load(open(f'{tabMFSEDdir}/MFSED_nB_{nB}_dimF_{dimF}.pkl', 'rb')) 
            for dQ in dQList:        
                tabQMFSED = np.round(tabMFSED/dQ).astype(int)            
                fct_args = zip(idSynSamples, repeat(synSamples), repeat(borders), repeat(tabQMFSED))

                pool = mp.Pool(32)            
                scoresSEDandSEDQ = pool.starmap(compSEDandSEDQ, fct_args)
                scoresSED, scoresSEDQ = zip(*scoresSEDandSEDQ)
                gc.collect()   
    
                scoresSED = np.array(scoresSED, dtype = float)
                scoresSEDQ = np.array(scoresSEDQ, dtype = int)
                        
                pool.close()
                pool.join()
                
                r = stats.pearsonr(scoresSEDQ, scoresSED)[0]
                pearson[dQ].append(r)
                                    
                r = np.round(r, 4)   
                plotTitle = f'dimF = {dimF} levels = 2^{nB} dQ = {dQ} Pearson = {r}'
                print(plotTitle)
                plotFile = f'{resDir}/Scatter_SEDvsMFSED_dimF_{dimF}_nB_{nB}_dQ_{dQ}_{timeTag}.pdf'
                plotScatter(plotFile, plotTitle, scoresSED, scoresSEDQ, 'SED', 'SED Quantized')    
    
        plotPearson(f'{resDir}/Pearson_SED_{dimF}.pdf', f'MFSED (dimF = {dimF})', nBList, pearson)
        pickle.dump((nBList, dQList, pearson), open(f'{resDir}/results_dimF_{dimF}_{timeTag}.pkl', 'wb'))


    
     
    
  




if __name__ == '__main__':
    main()



























