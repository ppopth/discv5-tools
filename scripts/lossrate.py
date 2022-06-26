#!/bin/python
import json
import matplotlib as mpl
import matplotlib.pyplot as plt
import matplotlib.ticker as mtick
import numpy as np
import sys

if len(sys.argv) != 3:
    print('usage: lossrate.py [measurement-result-file] [measurement-name]')
    sys.exit(1)

f = open(sys.argv[1])
nodes = json.loads(f.read())

lrs = []
for node in nodes:
    lrs.append(node['Result']['LossRate']*100)
fig, ax = plt.subplots()  # Create a figure containing a single axes.
ax.set_title(f'Packet loss rate distribution of {len(nodes)} nodes ({sys.argv[2]})')
ax.set_xlabel('Percent of packets are lost')
ax.xaxis.set_major_formatter(mtick.PercentFormatter())
ax.set_ylabel('number of nodes')
ax.hist(lrs, np.arange(0, 100, 1))
fig.savefig(f'{sys.argv[2]}-lossrate.png')
