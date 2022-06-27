# discv5-tools

The collection of tools used with discv5

| Tool            | Description |
|-----------------|-------------|
| [network-measure](#network-measure) | Used to measure the network property of nodes in the network |

## Building

Run `make` to build the whole project and all the binaries will be in `/bin`

## network-measure

*network-measure* is used to measure the network properties of the nodes in discv5 network. It can be used to measure an individual node or used to crawl the entire network and measure every node found.

Run the following command to measure an individual node specified by the ENR in the `-enr` option. The result is shown below the command showing the average RTT of 327.647ms and 0% of packet loss.
```
$ ./bin/network-measure -enr enr:-Ku4QHqVeJ8PPICcWk1vSn_XcSkjOkNiTg6Fmii5j6vUQgvzMc9L1goFnLKgXqBJspJjIsB91LTOleFmyWWrFVATGngBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhAMRHkWJc2VjcDI1NmsxoQKLVXFOhp2uX6jeT0DvvDpPcU8FWMjQdR4wMuORMhpX24N1ZHCCIyg
2022/06/27 14:32:39 started discv5-tools/network-measure
result: &{327.647004ms 0}
```

Run the following command to crawl the entire network and measure every node found.
```
$ ./bin/network-measure -crawl -file nodes.json
```
The option `-crawl` specifies that we want to crawl the network. The option `-file` specifies the file containing the previously crawled nodes in the network. This file may not exist if the command is run for the first time.

After the command is run, it will crawl the network indefinitely and measure the new nodes or re-measure the existing nodes if their new ENRs are found. The current set of the nodes is saved into the file specified in the `-file` option every minute.

At the same, every node in the set is checked every 15 minutes if it's still alive. If it's not, it's removed from the set.

### Log messages

Let's see the log messages for some specific node. Let's say for the node with id `cb4f66af34184cbe`.

```
2022/06/23 08:33:51 crawler: found alive node (id=cb4f66af34184cbe)
2022/06/23 08:36:29 nodeset: added id=cb4f66af34184cbe result={79.762657ms 0.5} nodeset={len=6016}
2022/06/23 08:51:29 nodeset: refreshed id=cb4f66af34184cbe nodeset={len=6038}
2022/06/23 09:06:30 nodeset: refreshed id=cb4f66af34184cbe nodeset={len=6053}
2022/06/23 09:21:36 nodeset: refreshed id=cb4f66af34184cbe nodeset={len=6076}
2022/06/23 09:36:43 nodeset: refreshed id=cb4f66af34184cbe nodeset={len=6086}
2022/06/23 09:51:43 nodeset: refreshed id=cb4f66af34184cbe nodeset={len=6102}
2022/06/23 09:52:31 crawler: found alive node (id=cb4f66af34184cbe)
2022/06/23 09:55:11 nodeset: updated id=cb4f66af34184cbe result={106.017457ms 0.5} nodeset={len=6101}
2022/06/23 09:56:04 crawler: found alive node (id=cb4f66af34184cbe)
2022/06/23 10:03:38 crawler: found alive node (id=cb4f66af34184cbe)
2022/06/23 10:10:23 nodeset: refreshed id=cb4f66af34184cbe nodeset={len=6134}
2022/06/23 10:18:02 crawler: found alive node (id=cb4f66af34184cbe)
2022/06/23 10:20:44 nodeset: updated id=cb4f66af34184cbe result={115.101079ms 0.5} nodeset={len=6156}
```

The log showed that the crawler found a new node at 08:33:51 and the measurement is finished and the node is added to the node set at 08:36:29 after that, every 15 minutes, it's checked if it's still alive. If so, the log shows that the node is refreshed. If not, the node will be removed like the log messages shown below.

You can see that the crawler found the node again at 09:52:31 and we found that the ENR found is newer than the one in the node set, so we re-measured the node and updated it in the node set.

The crawler found the node again at 09:56:04 and 10:03:38 but this time the ENR is not newer, so the node set is not updated.

```
2022/06/27 08:22:11 nodeset: refreshed id=a2121786c3182967 nodeset={len=6911}
2022/06/27 08:37:11 nodeset: refreshed id=a2121786c3182967 nodeset={len=6917}
2022/06/27 08:52:27 nodeset: removed id=a2121786c3182967 nodeset={len=6915}
```

### Nodes JSON file structure

```json
[
  {
    "NodeUrl": "enr:-Ly4QIXwKzBf1tb5rMjdIZa2NC9EcInj--VvzLsMVfENlrgBMILx73BGBT7auSi2NtSmAP21XSvh08MR11zcJNmxzPACh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCC9KcrAQAQIP__________gmlkgnY0gmlwhCKWcHWJc2VjcDI1NmsxoQKcjJu-2gO2DfY0UlYcgrUuid7l5_c9sL0N9rYfnRo-lohzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A",
    "Result": {
      "Rtt": 17952946,
      "LossRate": 0.51
    },
    "RefreshedAt": "2022-06-26T14:41:00.7755209Z",
    "UpdatedAt": "2022-06-22T22:41:21.593392691Z"
  },
  {
    "NodeUrl": "enr:-LO4QFkdG-i0Y8zXi-tl2IYbI1tonC_9VWHha3fyA6D07f91O6ddh1FTKVbq6CQ_sxvWck1y40FxFrBOpZBaDKixFM2B4odhdHRuZXRziAAAAQwAAAAAhGV0aDKQr8qroAEAAAD__________4JpZIJ2NIJpcIRGhdxLiXNlY3AyNTZrMaEDDIT6x60cpMIA0oc3ILoYYcxBGRVlZeukdjzkQzk5IXuDdGNwgiMog3VkcIIjKA",
    "Result": {
      "Rtt": 217691426,
      "LossRate": 0.02
    },
    "RefreshedAt": "2022-06-26T14:41:00.738764111Z",
    "UpdatedAt": "2022-06-26T11:10:38.534369743Z"
  }
]
```
The JSON file that stores the node set found by the crawler is an array of node objects. Each node object has four members: `NodeUrl`, `Result`, `RefreshedAt`, and `UpdatedAt`. **No two node objects have the same node ID.**

`NodeUrl` is the currently found ENR of the node. `RefreshedAt` is the timestamp of the last time the node is checked if it's alive. `UpdatedAt` is the timestamp that the node is found or the last time the ENR is updated.

`Result` is the result of the measurement. Currently there are only two things we measured: the RTT (measured as nanoseconds) and the packet loss rate (measured as $\frac{number\ of\ lost\ packets}{number\ of\ packets\ sent}$).

### Measurement

When a node is measured, we send 100 [ordinary message packets](https://github.com/ethereum/devp2p/blob/master/discv5/discv5-wire.md#ordinary-message-packet-flag--0) with random message data. Then the node is supposed to send a [WHOAREYOU packet](https://github.com/ethereum/devp2p/blob/master/discv5/discv5-wire.md#whoareyou-packet-flag--1) back. Note that both ordinary message packetes and WHOAREYOU packets are UDP packets, so there is no overhead in the transport layer.

If we don't receive a WHOAREYOU packet back after **3 seconds**, we count that round as a packet loss. If we do, we include its RTT to the accumluated average RTT.

After all 100 rounds of packets, we measure the average RTT as the average among all the successful rounds and the packet loss rate as the lost rounds divided by 100.

Notice that we decided to send ordinary message packets with random message data to measure the RTT, not [PING request](https://github.com/ethereum/devp2p/blob/master/discv5/discv5-wire.md#ping-request-0x01) or [FINDNODE request](https://github.com/ethereum/devp2p/blob/master/discv5/discv5-wire.md#findnode-request-0x03), because such requests require a handshake which significantly increases the latency.
