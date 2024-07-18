import numpy as np
from scipy import optimize, stats
from scipy.integrate import quad
import scipy.special as sc
import glob

from utils import *

RANDN = 32

def getIndexQF(x, t):
    return len(np.where(t<x)[0])

'''
        Updated computeScore function for BIP.

        We're going to have x and y positions for both the table and the rand.
'''

def computeScore(x, y, t, table, rand):
    nFeat = len(x)
    score = 0
    for i in range(nFeat):
        ix = getIndexQF(x[i], t)
        iy = getIndexQF(y[i], t)
        score += (table[ix,iy] + rand[ix,iy])
    return score


def compIPandIPQ(i, synSamples, t, tabQMFIP, tabRand):
    x = synSamples[:,i]
    y = synSamples[:,i+1]
    ip = np.sum(x*y)
    ipQ = computeScore(x, y, t, tabQMFIP, tabRand)
    return ip, ipQ



# x is a feature element in [-1,1] from a normalized feature vector
# m is the dimension of the feature vector

def unifOnHyperSpherePDF(x, m):
    c = 1/sc.beta(0.5, (m-1)/2)
    p = c * np.power(1-x**2, (m-3)/2)
    return p

def unifOnHyperSphereXExp(x, m):
    return x*unifOnHyperSpherePDF(x, m)


def genTabMFIP(nB, featDim):
    t = findEquiProbThresholdsBrentq(2**nB , featDim)
    tabMFIP = jointUnifOnHyperSphereConditionalExp(nB, featDim, t)
    return t, tabMFIP


def genBordersLookupTables(nB, featDim):
    t = findEquiProbThresholdsBrentq(2**nB , featDim)
    tabMFIP = jointUnifOnHyperSphereConditionalExp(nB, featDim, t)
    #tabMFSED = jointUnifOnHyperSphereConditionalExpSED(nB, featDim, t)
    return t, tabMFIP#, tabMFSED

# Equation (21) with the assumption that m is even

def unifOnHyperSphereCDF(x, m):
    c = 1/sc.beta(0.5, (m-1)/2)
    cdf = 0
    n = m-2
    firstTermX = (0.5**n) * sc.binom(n, n/2) * np.arcsin(x)
    firstTermMinus1 = (0.5**n) * sc.binom(n, n/2) * np.arcsin(-1)
    secondTermX = 0
    sumBoundK = int(n/2) - 1
    for k in range(sumBoundK+1): # +1 to make sure sumBoundK is included
        sumOverJ = 0
        sumBoundJ = n-2*k
        internalCoeff = sc.binom(n, k) / sumBoundJ
        jList = np.arange(1, sumBoundJ, 2)
        for j in jList:
            fristTermJ = sc.binom(sumBoundJ, j) * ((-1)**((j-1)/2))
            a = (np.sqrt(1-x**2))**(sumBoundJ-j)
            b = x**j
            secondTermJ = a * b
            sumOverJ += fristTermJ * secondTermJ
        secondTermX +=  internalCoeff  * sumOverJ
    secondTermX = secondTermX / (2**(n-1))
    cdf = c * (firstTermX + secondTermX - firstTermMinus1)
    return cdf


def findEquiProbThresholdsBrentq(n, m):
    borders = []
    p = np.arange(1,n)/n
    for pi in p:
        def cdf(x):
            return unifOnHyperSphereCDF(x, m)-pi
        root = optimize.brentq(cdf, -1, 1)
        borders.append(root)
    return borders


def jointUnifOnHyperSphereConditionalExp(nB, featDim, t):
    inf = 1
    nL = 2**nB # cellProba = 1/(nL**2)
    nL2 = int(nL/2)
    # t = findEquiProbThresholds(nL, featDim)
    pSame = np.zeros((nL, nL))
    # compute the bins that are on the diag and diag_inver of pSame with mirroring
    for i in range(nL2):
        if (i == 0):
            xmin = -inf
            xmax = t[0]
            ymin1 = xmin
            ymax1 = xmax
            ymin2 = t[-1]
            ymax2 = inf
        else:
            xmin = t[i-1]
            xmax = t[i]
            ymin1 = xmin
            ymax1 = xmax
            ymin2 = t[-i-1]
            ymax2 = t[-i]

        xE = quad(lambda x: unifOnHyperSphereXExp(x, featDim), xmin, xmax)[0]
        yE = quad(lambda x: unifOnHyperSphereXExp(x, featDim), ymin1, ymax1)[0]
        pSame[i,i] = (xE*yE)*(nL**2)
        pSame[nL-1-i,nL-1-i] = pSame[i,i]
        yE = quad(lambda x: unifOnHyperSphereXExp(x, featDim), ymin2, ymax2)[0]
        pSame[i,nL-1-i] = (xE*yE)*(nL**2)
        pSame[nL-1-i,i] = pSame[i,nL-1-i]
    # compute the bin out of both diags of pSame with mirroring
    for i in range(1,nL-1):
        xmin = t[i-1]
        xmax = t[i]
        for j in range(min(nL-i-1, i)):
            if (j == 0):
                ymin = -inf
                ymax = t[0]
            else:
                ymin = t[j-1]
                ymax = t[j]
            xE = quad(lambda x: unifOnHyperSphereXExp(x, featDim), xmin, xmax)[0]
            yE = quad(lambda x: unifOnHyperSphereXExp(x, featDim), ymin, ymax)[0]
            pSame[i,j] = (xE*yE)*(nL**2)
            pSame[j,i] = pSame[i,j]
            pSame[nL-1-i, nL-1-j] = pSame[i,j]
            pSame[nL-1-j,nL-1-i] = pSame[i,j]

    return pSame


def matedIP(dataDir, subject, t, tabQIP, tabRand, nFix = 15):
    matedScIP = []
    matedScIPQ = []
    subjectImgs = glob.glob(dataDir+f'/{subject}/*.npy')
    subjectImgs.sort()

    n = len(subjectImgs)
    subjectImgs = subjectImgs[:n] if (n < nFix) else subjectImgs[:nFix]

    indexes = [(i,j) for i in range(len(subjectImgs)) for j in range(len(subjectImgs)) if i<j]
    for (i,j) in indexes:
        x = np.load(subjectImgs[i])
        x = normalizeSample(x)
        y = np.load(subjectImgs[j])
        y = normalizeSample(y)
        matedScIP.append(np.sum(x*y))
        matedScIPQ.append(computeScore(x, y, t, tabQIP, tabRand))
    return matedScIP, matedScIPQ


def nonMatedIP(dataDir, subject, subjectIDs, t, tabQIP, tabRand, nFix = 5):

    nonMatedScIP = []
    nonMatedScIPQ = []


    subjID = subjectIDs.index(subject)
    subjectList = subjectIDs[subjID+1:]


    fixFR = glob.glob(dataDir+f'/{subject}/*.npy')
    fixFR.sort()
    n = len(fixFR)

    fixFR = fixFR[:n] if (n < nFix) else fixFR[:nFix]


    impFR = [ dataDir+f'/{subj}' for subj in subjectList]

    for tempFR in fixFR:
        x = np.load(tempFR)
        x = normalizeSample(x)


        for probeFR in impFR:
            embProbFR = glob.glob(probeFR+'/*.npy')[0]
            y = np.load(embProbFR)
            y = normalizeSample(y)

            nonMatedScIP.append(np.sum(x*y))
            nonMatedScIPQ.append(computeScore(x, y, t, tabQIP, tabRand))

    return nonMatedScIP, nonMatedScIPQ

