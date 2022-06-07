package crawler

import (
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

type nodeInfo struct {
	url string
}

var (
	nodeInfos = []nodeInfo{
		{url: "enr:-KO4QDBsHwuYdxyb_KR_sJEt-5ikIsdfyQHK6zi72KiDXTIgDGf9mQl8hen6ycgbJyaSgjbe9_lLy6lcZZA5iwECoCWCATWEZXRoMpCvyqugAQAAAP__________gmlkgnY0gmlwhAMTwp2Jc2VjcDI1NmsxoQOGl6EENtmMz8v16Tr31ju-FQn54B0zJBb8WKXnbZjR84N0Y3CCIyiDdWRwgiMo"},
		{url: "enr:-Ku4QLylXZ0DWTelCTZQJxl2lsJFYYNk9B_Q2YXYfnxAiYCsRyOJnbVvxWRnQqiD1KTpa4YCdPwcdilx0ALtjIwLRjIHh2F0dG5ldHOIAAAAAAAAAACEZXRoMpC1MD8qAAAAAP__________gmlkgnY0gmlwhDayLMaJc2VjcDI1NmsxoQK2sBOLGcUb4AwuYzFuAVCaNHA-dy24UuEKkeFNgCVCsIN1ZHCCIyg"},
		{url: "enr:-Ly4QKQ4BqHAOloSz-_lYVbfPpuAbn3uFxFiRSmWNzSEJZrsVnG-kTqjAleCu-KkSxvmIpt_ZIMmgUMbrWGdvDyEuM08h2F0dG5ldHOI__________-EZXRoMpDucelzYgAAcf__________gmlkgnY0gmlwhES3XM2Jc2VjcDI1NmsxoQK79EwWY2Zi9wvUKcFGkN3-VwoMvLLCJCKHQxFH6xgPyYhzeW5jbmV0cw-DdGNwgiMog3VkcIIjKA"},         // has attnets
		{url: "enr:-MK4QB-ycOj1GuzRW8pjXiMxRhQz0Yby-Z_KWwZ_D3ddGy2dbWOVTmj3E6hFkoFTGeey1qJhq2bddsSnMz9xvWNneKGGAX26RhZSh2F0dG5ldHOI-_-___7__9-EZXRoMpCvyqugAQAAAP__________gmlkgnY0gmlwhCPc-ZyJc2VjcDI1NmsxoQMJdU5g6WmwFY10zH2rB7qyM-3hBgPH9mTtRLu-zv1FIYhzeW5jbmV0cwCDdGNwgjLIg3VkcIIu4A"}, // has attnets
		{url: "enr:-KO4QDBsHwuYdxyb_KR_sJEt-5ikIsdfyQHK6zi72KiDXTIgDGf9mQl8hen6ycgbJyaSgjbe9_lLy6lcZZA5iwECoCWCATWEZXRoMpCvyqugAQAAAP__________gmlkgnY0gmlwhAMTwp2Jc2VjcDI1NmsxoQOGl6EENtmMz8v16Tr31ju-FQn54B0zJBb8WKXnbZjR84N0Y3CCIyiDdWRwgiMo"},
		{url: "enr:-IS4QDAyibHCzYZmIYZCjXwU9BqpotWmv2BsFlIq1V31BwDDMJPFEbox1ijT5c2Ou3kvieOKejxuaCqIcjxBjJ_3j_cBgmlkgnY0gmlwhAMaHiCJc2VjcDI1NmsxoQJIdpj_foZ02MXz4It8xKD7yUHTBx7lVFn3oeRP21KRV4N1ZHCCIyg"},
	}
)

func TestNew(t *testing.T) {
	bootNodes := make([]*enode.Node, 0, len(nodeInfos))
	for _, info := range nodeInfos {
		bootNodes = append(bootNodes, enode.MustParse(info.url))
	}
	c := &Config{BootNodes: bootNodes}
	n := New(c)
	if n == nil {
		t.Error("New returns nil")
	}
	if len(n.config.BootNodes) != len(bootNodes) {
		t.Errorf("New changes boot nodes")
	}
	if n.log == nil {
		t.Error("New doesn't create the logger")
	}
	if n.config != c {
		t.Error("New doesn't reference the config")
	}
}
