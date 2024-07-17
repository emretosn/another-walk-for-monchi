import numpy as np
import os

labels = np.load('LFW_labels.npy', mmap_mode='r')
embeddings = np.load('LFW_embeddings.npy', mmap_mode='r')

#if len(labels) == len(embeddings):
#    for i in range(len(labels)):
#        print(labels[i])
#        print(embeddings[i])

#print(len(set(labels)))
#print(len(labels))

d = {}
for person in labels:
    if person in d:
        d[person] += 1
    else:
        d[person] = 1

for k, v in d.items():
    print(k, v)

#for i in range(len(labels)):
#    label_dir = f'LFW/{labels[i]}'
#    if os.path.isdir(label_dir):
#        j = len([entry for entry in os.listdir(label_dir) if os.path.isfile(os.path.join(label_dir, entry))])
#    else:
#        os.mkdir(label_dir)
#        j = 0
#    file_path = os.path.join(label_dir, f'{j}.npy')
#    np.save(file_path, embeddings[i])
