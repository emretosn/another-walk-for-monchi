
import pandas as pd
import numpy as np
import seaborn as sns
import matplotlib.pyplot as plt
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







# plots

def plotPearson(filePLT,title, nBList, pearson, labelX='Bits', labelY='Pearson'):
    sns.set(style = "darkgrid")
    df = pd.DataFrame(pearson)
    df = df.set_index(nBList)
    #df = df.drop([0.1], axis=1)
    fig, ax = plt.subplots()
    sns.lineplot(data=df)
    ax.set(xlabel=labelX, ylabel=labelY)
    fig.suptitle(title)
    plt.subplots_adjust(left=0.2, bottom=0.2)
    plt.savefig(filePLT)

def plotScatter(filePLT, title, scoresIP, scoresIPQ, labelX, labelY):
    sns.set(style = "darkgrid")
    df = pd.DataFrame({"IP": scoresIP,"IPQ": scoresIPQ})
    fig, ax = plt.subplots()
    sns.scatterplot(data=df, x='IP', y='IPQ')
    ax.set(xlabel=labelX, ylabel=labelY)
    fig.suptitle(title)
    plt.subplots_adjust(left=0.2, bottom=0.2)
    plt.savefig(filePLT)

def get_fmr_tmr_Prec(matedScores, nonMatedScores, precision = 0.1):
    """Calculates the FMR and TMR from mated and nonMated scores
            Parameters
            ----------
            matedScores :   array-like of shape = (n_genuine,)
                            mated scores resulting from genuine comparison of two same-subject-samples
            nonMatedScores : array-like of shape = (n_impostor,)  
                            nonMated scores resulting from impostor comparison of two different-subject-samples
            precision :     step between two adjacent threshold scores
                            equals to 1 when the scores are integers and 1E-n when they are real-values 
            
            
            Returns
            -------
            fmr :   array-like representing the False Match Rate
                    the rate of impostor scores as a function of the threshold
            tmr :   array-like representing the True Match Rate
                    the rate of genuine scores as a function of the threshold
            scores : array-like representing all possible thresholds
    """
    # determine all possible thresholds
    minS = min(min(matedScores), min(nonMatedScores))
    maxS = max(max(matedScores), max(nonMatedScores))
    scores = np.arange(minS, maxS + precision, precision)
    lenMated = len(matedScores)
    lennonMated = len(nonMatedScores)
    # calculate the occurrence of mated scores 
    pgen = np.histogram(matedScores, bins=scores, density=False)[0]
    # calculate the genuine probability 
    pgen = pgen/lenMated
    # calculate the TMR as a function of the thresholds (that is scores)
    tmr = np.maximum(0, 1 - np.cumsum(pgen))
    # calculate the occurrence of nonMated scores 
    pimp = np.histogram(nonMatedScores, bins=scores, density=False)[0]
    # calculate the impostor probability
    pimp = pimp/lennonMated
    # calculate the FMR as a function of the thresholds (that is scores)    
    fmr = np.maximum(0, 1 - np.cumsum(pimp))  
    # for a threshold equals to scores[i], FMR equals to fmr[i] and TMR equals to tmr[i] 
    return scores, fmr, tmr


def calc_eerPT(fmr, tmr):
    """Calculates the Equal Error Rate point from fmr and 1-tmr
            Parameters
            ----------
            fmr :   array-like representing the False Match Rate
                    the rate of impostor scores as a function of the threshold
            tmr :   array-like representing the True Match Rate
                    the rate of genuine scores as a function of the threshold
            
            Returns
            -------
            eer : point where fmr and fnmr (1-tmr) are equal
    """
    fnmr = 1-tmr
    x = fnmr - fmr
    if ((fmr.size == 0) or (fnmr.size == 0)):
        return np.inf
    
    index = np.argmin(np.abs(x))   
    
    if (index == 0):
        return (fnmr[index] + fmr[index])/2 

    if (index == len(fmr)):
        return (fnmr[index] + fmr[index])/2 

    if (fmr[index] <= fnmr[index]):
        l_index = index - 1
        r_index = index
    else:
        l_index = index
        r_index = index + 1

    d_1 = fmr[l_index] - fnmr[l_index]
    d_2 = fmr[r_index] - fnmr[r_index]

    if (d_1 - d_2) == 0:
        s_val = 0.5
    else:
        s_val = d_1 / (d_1 - d_2)

    eer = fnmr[l_index] + s_val*(fnmr[r_index] - fnmr[l_index])
    return eer


def plot_MultipleDETsSaved(plotFile, plotTitle, fmrList, tmrList, labelList, linewidthList):
    """Plots and saves the DET curves of fmrList and 1-tmrList
            Parameters
            ----------
            plotFile : path to where the plot should be saved
            plotTitle : path to  
            fmrList : list of FMRs 
            tmrList : list of TMRs  
            labelList : list of labels corresponding to the DET curve of the i-th fmrList and 1-tmrList
            linewidthList : list of integers specifying the linewidth of the i-th DET curve
            
            Returns
            -------
            
    """
    eer_line = np.logspace(-4,0,100) 
    fig, ax = plt.subplots()
    for fmr, tmr, lab, lw in zip(fmrList, tmrList, labelList, linewidthList):
        # fnmr is equal to 1-tmr
        ax.loglog(fmr, 1-tmr, label = lab, linewidth=lw)
    ax.loglog(eer_line, eer_line, label = 'EER')
    ax.set_aspect('equal')
    ax.set_xlim([1E-4, 1])
    ax.set_ylim([1E-4, 1])
    ax.set(xlabel='FMR', ylabel='FNMR')
    ax.legend()
    fig.suptitle(plotTitle)
    plt.savefig(plotFile)






