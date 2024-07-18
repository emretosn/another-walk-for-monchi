import pickle
import os

from utils import *
from mfbrTabs import *


seed = 4521
numSamples = 2*100000 

print('Normalized samples\nnumSamples = ', numSamples)

pathSynthetic = f'./data/Synthetic/'
os.makedirs(pathSynthetic, exist_ok=True)

dimFlist = [32, 64, 128, 256, 512]
for dimF in dimFlist:
    synSamples = genSynSamplesNormalDist(seed, numSamples, dimF)    
    pickle.dump((synSamples), open('./data/Synthetic/syntheticSamples_dimF_{}.pkl'.format(dimF), "wb"))    
    print('Synthetic Samples dimF = {}'.format(dimF))









