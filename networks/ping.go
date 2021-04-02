package networks

import (
	"time"

	"github.com/go-ping/ping"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

func pingNode() (bestNode string) {
	nodes := [5]string{"fr1.vtconline.org", "p2proxy.vertcoin.org", "p2p-usa.xyz", "p2p-ekb.xyz", "173.198.248.34"}
	results := [len(nodes)]time.Duration{}

	bestNode = nodes[0]
	lowest := results[0]

	for i := 0; i < len(nodes); i++ {
		pinger, err := ping.NewPinger(nodes[i])
		pinger.SetPrivileged(true)  //This line is needed for windows because of ICMP
		pinger.Timeout = 5000000000 //If response time is longer than 5 seconds, the pinger will exit regardless of how many packets have been recieved
		if err != nil {
			logging.Warn("Error: Check if you are connected to the internet")
			logging.Warn(err)
		}
		pinger.Count = 3
		err = pinger.Run()
		if err != nil {
			logging.Warn("Error: Check if you are connected to the internet")
			logging.Warn(err)
		}
		results[i] = pinger.Statistics().AvgRtt
		logging.Infof("%s: %v \n", nodes[i], results[i])

		if results[i] < lowest && results[i] != 0 || lowest == 0 {
			bestNode = nodes[i]
			lowest = results[i]
		}
	}
	if bestNode == "p2proxy.vertcoin.org" { //Currently this node uses port 9172 instead of 9171, if this changes this statement can be removed and 9171 can be added to all nodes
		bestNode += ":9172"
	} else {
		bestNode += ":9171"
	}

	logging.Infof("Selected node: %s\n", bestNode)
	return bestNode
}

var selectedNode string = pingNode()

func getClosestNodeStratum() (stratum string) {
	stratum = "stratum+tcp://"
	stratum += selectedNode
	return stratum
}

func getClosestNodeURL() (URL string) {
	URL = "http://"
	URL += selectedNode
	URL += "/"
	return URL
}
