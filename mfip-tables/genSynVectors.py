import pickle
import pandas as pd
import os

from utils import *
from mfbrTabs import *


seed = 4521
numSamples = 2*1000

print('Normalized samples\nnumSamples = ', numSamples)

pathSynthetic = f'./data/Synthetic/'
os.makedirs(pathSynthetic, exist_ok=True)

dimFlist = [512] #[32, 64, 128, 256, 512]
for dimF in dimFlist:
    synSamples = genSynSamplesNormalDist(seed, numSamples, dimF)
    df = pd.DataFrame(synSamples)
    df.to_csv('./data/Synthetic/syntheticSamples_dimF_{}.csv'.format(dimF), index=False)
    print('Synthetic Samples dimF = {}'.format(dimF))

