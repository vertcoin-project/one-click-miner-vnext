package ping

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/go-ping/ping"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

type Conditions struct {
	MaxFee      float64
	MaxMiners   int
	MaxNextPing time.Duration
	PingPackets int
	PingTimeout time.Duration
}

var Set = Conditions{
	MaxFee:      2,
	MaxMiners:   35,
	MaxNextPing: 25 * time.Millisecond, //Prevents the miner from using a node with too high latency compared to other nodes. Should prevent the miner from using nodes outside of their region
	PingPackets: 2,
	PingTimeout: 2000000000,
}

type SelectedNode struct {
	P2PoolStratum string
	P2PoolURL     string
}

var Selected SelectedNode

func GetSelectedNode(testnet bool) {
	if testnet {
		Selected = SelectedNode{
			P2PoolStratum: networks.Active.P2ProxyStratum,
			P2PoolURL:     networks.Active.P2ProxyURL,
		}
	} else {
		selector()
	}
}

func selector() {
	_, err := GetNodeInformation("http://127.0.0.1:9171/")
	if err == nil {
		logging.Infof("Selected local p2pool node\n")
		Selected = SelectedNode{
			P2PoolStratum: "stratum+tcp://127.0.0.1:9171",
			P2PoolURL:     "http://127.0.0.1:9171/",
		}
	} else {
		logging.Infof("No local node detected, selecting other public nodes\n")

		err = PingNodes()
		if err != nil {
			logging.Warnf("Nodes could not be pinged, selecting random node\n")
			rand.Seed(time.Now().Unix())
			randInt := rand.Intn(len(NodeList))
			Selected = SelectedNode{
				P2PoolStratum: NodeList[randInt].Stratum,
				P2PoolURL:     NodeList[randInt].URL,
			}
			logging.Infof("%s has been randomly selected\n", NodeList[randInt].Hostname)
		} else {

			sort.Slice(NodeList, func(i, j int) bool {
				return NodeList[i].PingTime < NodeList[j].PingTime
			})

			for i := 0; i < len(NodeList); i++ {
				nodeInformation, _ := GetNodeInformation(NodeList[i].URL)
				fee := CheckFee(nodeInformation)
				if fee {
					currentMiners := CheckCurrentMiners(nodeInformation)
					if currentMiners {
						Selected = SelectedNode{
							P2PoolStratum: NodeList[i].Stratum,
							P2PoolURL:     NodeList[i].URL,
						}
						logging.Infof("%s selected, fulfilled all requirements\n", NodeList[i].Hostname)
						break
					} else { //If the node fulfills the max fee requirement but there is more than the set MaxNextping to the next node it will select the current
						determineNextPingTime := NodeList[i+1].PingTime
						determineNextPingTime -= NodeList[i].PingTime
						if determineNextPingTime > Set.MaxNextPing {
							Selected = SelectedNode{
								P2PoolStratum: NodeList[i].Stratum,
								P2PoolURL:     NodeList[i].URL,
							}
							logging.Infof("%s selected, next node has too high ping time\n", NodeList[i].Hostname)
							break
						}
						logging.Warnf("%s had more than %v miners, trying new inorder to retain efficiency\n", NodeList[i].Hostname, Set.MaxMiners)
					}
				} else {
					logging.Warnf("%s is either unreachable or had more than a %f fee, trying new\n", NodeList[i].Hostname, Set.MaxFee)
				}
			}
		}
	}
}

func PingNodes() error {
	for i := 0; i < len(NodeList); i++ {
		pinger, err := ping.NewPinger(NodeList[i].Hostname)
		pinger.SetPrivileged(true)       //This line is needed for windows because of ICMP
		pinger.Timeout = Set.PingTimeout //Sets the time for which the pinger will timeout regardless of how many packets there has been recieved
		if err != nil {
			logging.Warn("Error: Check if you are connected to the internet")
			logging.Warn(err)
			return err
		}
		pinger.Count = Set.PingPackets //Number of packets to be sent to each node
		err = pinger.Run()
		if err != nil {
			logging.Warn("Error: Check if you are connected to the internet")
			logging.Warn(err)
			return err
		}
		NodeList[i].PingTime = pinger.Statistics().AvgRtt
		logging.Infof("%s: %v \n", NodeList[i].Hostname, NodeList[i].PingTime)
	}
	return nil
}

//Instead of making a http request to the node each time we need to get information, we do it once and then reuse the collected data.
func GetNodeInformation(NodeURL string) (jsonPayload map[string]interface{}, err error) {
	err = util.GetJson(fmt.Sprintf("%slocal_stats", NodeURL), &jsonPayload)
	if err != nil {
		logging.Errorf("Unable to fetch node information\n", err.Error())
		return jsonPayload, err
	}
	return jsonPayload, nil
}

func CheckFee(jsonPayload map[string]interface{}) bool {
	fee, ok := jsonPayload["fee"].(float64)
	if !ok {
		return false
	}
	donationFee, ok := jsonPayload["donation_proportion"].(float64)
	if !ok {
		return false
	}
	fee += donationFee
	if fee > Set.MaxFee {
		return false
	}
	return true
}

//To ensure efficiency of the selected p2pool node a limit of miners has been put in place, returns true if the number is equal to Maxminers or below
func CheckCurrentMiners(jsonPayload map[string]interface{}) bool {
	currentMiners, _ := jsonPayload["miner_hash_rates"].(string)
	if len(currentMiners) > Set.MaxMiners {
		return false
	}
	return true
}
