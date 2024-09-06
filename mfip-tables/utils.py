
import numpy as np
from numpy import linalg


def normalizeSample(sample):
    return sample/linalg.norm(sample)

def normalizeData(dataSample):
    dFeat, nSamples = dataSample.shape
    normS = dataSample**2
    normS = np.sqrt(np.sum(normS, axis = 0))
    sampNorm = np.zeros((dFeat, nSamples))
    for i, norml in enumerate(normS):
        sampNorm[:,i] = dataSample[:,i]/norml
    return sampNorm

def genSynSamplesNormalDist(seed, numSamples, dimVect):
    rng = np.random.default_rng(seed=seed)
    y = rng.normal(loc=0.0, scale=1.0, size=(dimVect, numSamples))
    return normalizeData(y)


def flattenList(lists):
    flatList = []
    for l in lists:
        flatList += l
    return flatList
