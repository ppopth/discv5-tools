#!/bin/python
import json
import matplotlib as mpl
import matplotlib.pyplot as plt
import numpy as np
import sys

if len(sys.argv) != 3:
    print('usage: latency.py [measurement-result-file] [measurement-name]')
    sys.exit(1)

f = open(sys.argv[1])
nodes = json.loads(f.read())

rtts = []
for node in nodes:
    rtts.append(node['Result']['Rtt']/10.0**9)
fig, ax = plt.subplots()  # Create a figure containing a single axes.
ax.set_title(f'Latency distribution of {len(nodes)} nodes ({sys.argv[2]})')
ax.set_xlabel('latency in second, 5ms/bin')
ax.set_ylabel('number of nodes')
ax.hist(rtts, np.arange(0, 0.5, 0.005))
fig.savefig(f'{sys.argv[2]}-latency.png')
