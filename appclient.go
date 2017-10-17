package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	//pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var _LOGGER = log.New(os.Stderr, "HLFAppCleintLogger:", log.Lshortfile)

type HLFApplicationClient struct {
	Client          fab.FabricClient
	Channel         fab.Channel
	ChannelClient   apitxn.ChannelClient
	EventHubMap     map[string]fab.EventHub
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
	//AdminUser       ca.User
	//OrdererAdmin ca.User
	ApplicationUser ca.User
	SDK             *deffab.FabricSDK
	PeerMap         map[string]fab.Peer
}

// Initialize reads configuration from file and sets up client, channel and event hub
func (sdkClient *HLFApplicationClient) Initialize() error {
	_LOGGER.Println("SDKClient.Initialize: start")

	// Create SDK setup for the integration tests
	sdkOptions := deffab.Options{
		ConfigFile: sdkClient.ConfigFile,
		//		OrgID:      setup.OrgID,
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/enroll_user",
		},
	}

	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return fmt.Errorf("Error initializing SDK: %s", err)
	}
	_LOGGER.Println("SDK Created ")
	sdkClient.SDK = sdk
	session, err := sdk.NewPreEnrolledUserSession(sdkClient.OrgID, "User1")
	if err != nil {
		return fmt.Errorf("Error getting admin user session for org: %s", err)
	}
	sc, err := sdk.NewSystemClient(session)
	if err != nil {
		return fmt.Errorf("NewSystemClient returned error: %v", err)
	}

	sdkClient.Client = sc
	//sdkClient.AdminUser = session.Identity()
	sdkClient.ApplicationUser = session.Identity()
	_LOGGER.Println("Application user context found ")
	/*ordererAdmin, err := sdk.NewPreEnrolledUser("ordererorg", "Admin")
	if err != nil {
		return fmt.Errorf("Error getting orderer admin user: %v", err)
	}
	*/
	/*sdkClient.OrdererAdmin = ordererAdmin
	_LOGGER.Println("Orderer admin is retrieved  ")
	*/
	chClient, err := sdk.NewChannelClient(sdkClient.ChannelID, "User1")
	if err != nil {
		return fmt.Errorf("Error in creating channel client: %v", err)
	}
	sdkClient.ChannelClient = chClient
	_LOGGER.Println("Chanel Client is setup  ")
	sdkClient.setUpEventHubMap()
	sdkClient.Initialized = true

	return nil
}

// InitConfig ...
func (sdkClient *HLFApplicationClient) InitConfig() (apiconfig.Config, error) {
	configImpl, err := config.InitConfig(sdkClient.ConfigFile)
	if err != nil {
		return nil, err
	}
	return configImpl, nil
}
func (sdkClient *HLFApplicationClient) setUpEventHubMap() (map[string]fab.EventHub, error) {
	if sdkClient.EventHubMap != nil {
		return sdkClient.EventHubMap, nil
	}
	peerConfig, err := sdkClient.Client.Config().PeersConfig(sdkClient.OrgID)
	if err != nil {
		return nil, fmt.Errorf("PeersConfig failed %v ", err)
	}
	eventHubMap := make(map[string]fab.EventHub, len(peerConfig))
	for _, p := range peerConfig {

		if p.URL != "" {
			eventHub, err := events.NewEventHub(sdkClient.Client)
			if err != nil {
				return nil, fmt.Errorf("NewEventHub failed %v", err)
			}
			_LOGGER.Printf("EventHub connect to peer (%s)", p.EventURL)
			serverHostOverride := ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			eventHub.SetPeerAddr(p.EventURL, p.TLSCACerts.Path, serverHostOverride)
			eventHubMap[serverHostOverride] = eventHub
		}
	}

	sdkClient.EventHubMap = eventHubMap

	return sdkClient.EventHubMap, nil
}

// GetChannel initializes and returns a channel based on config
func (sdkClient *HLFApplicationClient) GetChannel(channelID string, orgs []string) (fab.Channel, error) {
	client := sdkClient.Client
	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}

	ordererConfig, err := client.Config().RandomOrdererConfig()
	if err != nil {
		return nil, fmt.Errorf("RandomOrdererConfig() return error: %s", err)
	}
	serverHostOverride := ""
	if str, ok := ordererConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}
	orderer, err := orderer.NewOrderer(ordererConfig.URL, ordererConfig.TLSCACerts.Path, serverHostOverride, client.Config())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}

	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	for _, org := range orgs {

		peerConfig, err := client.Config().PeersConfig(org)
		if err != nil {
			return nil, fmt.Errorf("Error reading peer config: %v", err)
		}
		peerMap := make(map[string]fab.Peer, len(peerConfig))

		for _, p := range peerConfig {
			fmt.Println("Parsing peer ", p.EventURL)

			serverHostOverride = ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			endorser, err := deffab.NewPeer(p.URL, p.TLSCACerts.Path, serverHostOverride, client.Config())
			if err != nil {
				return nil, fmt.Errorf("NewPeer return error: %v", err)
			}
			peerMap[serverHostOverride] = endorser
			err = channel.AddPeer(endorser)
			if err != nil {
				return nil, fmt.Errorf("Error adding peer: %v", err)
			}

		}
		sdkClient.PeerMap = peerMap
	}
	sdkClient.Channel = channel
	return channel, nil
}
func (sdkClient *HLFApplicationClient) GenerateStatistics(peer string, txnIds []string) {
	blockInfo, err := sdkClient.Channel.QueryInfo()
	if err != nil {
		_LOGGER.Fatalf("Channel.QueryInfo() Failed %v \n", err)
		return
	}
	_LOGGER.Printf("Block info %s\n", blockInfo.String())
	block, err := sdkClient.Channel.QueryBlock(int(blockInfo.Height - 1))
	if err != nil {
		_LOGGER.Fatalf("Channel.QueryBlock() Failed %v \n", err)
		return
	}
	_LOGGER.Printf("Block details Header %s\n", block.Header.String())
	str := string(block.Data.GetData()[0])

	_LOGGER.Printf("Block details Data Line %d %s\n", 0, str)
	for _, line := range strings.Split(str, "\n") {
		_LOGGER.Printf("Line details %v", line)
	}
	previousBlock, err := sdkClient.Channel.QueryBlockByHash(blockInfo.CurrentBlockHash)
	if err != nil {
		_LOGGER.Fatalf("Channel.QueryBlockByHash() Failed %v \n", err)
		return
	}
	_LOGGER.Printf("Previous block details %v\n", previousBlock)
	for _, trxnID := range txnIds {
		trxnDetails, err := sdkClient.Channel.QueryTransaction(trxnID)
		if err == nil {
			_LOGGER.Printf("Trxn  details %s %s", trxnID, trxnDetails.String())
		}

	}
}

//InvokeQueryTranaction : TODO return results etc
func (sdkClient *HLFApplicationClient) InvokeQueryTranaction(chainCodeID string,
	fcn string, args [][]byte, peers []string) ([][]byte, error) {
	channel := sdkClient.Channel
	targets := []apitxn.ProposalProcessor{sdkClient.PeerMap[peers[0]]}
	ptx, _ := sdkClient.Client.NewTxnID()
	request := apitxn.ChaincodeInvokeRequest{
		Targets:     targets,
		Fcn:         fcn,
		Args:        args,
		ChaincodeID: chainCodeID,
		TxnID:       ptx,
	}
	_LOGGER.Printf("Pregenerated trxn Id %v", ptx)
	peerResponses, _, err := channel.SendTransactionProposal(request)
	if err != nil {
		_LOGGER.Fatalf("Chain code query failed with %v\n", err)
		return nil, err
	}
	respBytes := make([][]byte, len(peerResponses))
	for index, peerResponse := range peerResponses {
		respBytes[index] = peerResponse.TransactionProposalResult.ProposalResponse.Response.Payload
		_LOGGER.Printf("Query returned %s\n", respBytes[index])

	}
	return respBytes, nil
}

//InvokeTransaction : TODO : Return results etc
func (sdkClient *HLFApplicationClient) InvokeTransaction(chainCodeID string,
	fcn string, args [][]byte, peers []string) (string, error) {
	channel := sdkClient.Channel
	targets := []apitxn.ProposalProcessor{sdkClient.PeerMap[peers[0]]}
	ptx, _ := sdkClient.Client.NewTxnID()

	request := apitxn.ChaincodeInvokeRequest{
		Targets:     targets,
		Fcn:         fcn,
		Args:        args,
		ChaincodeID: chainCodeID,
		TxnID:       ptx,
	}
	_LOGGER.Printf("Pregenerated trxn Id %v", ptx)
	transactionProposalResponses, txnID, err := channel.SendTransactionProposal(request)

	if err != nil {
		_LOGGER.Fatalf("Error from the sendong trxn proposal %v", err)
		return txnID.ID, err
	}

	for _, v := range transactionProposalResponses {
		_LOGGER.Printf(" Endorser response %v \n", v.Status)
		if v.Err != nil {
			_LOGGER.Fatalf("Error from the sendong trxn proposal endorsers %v", err)
			return txnID.ID, err
		}
	}
	_LOGGER.Printf("Transaction id %v", txnID)

	tx, err := channel.CreateTransaction(transactionProposalResponses)
	if err != nil {
		_LOGGER.Fatalf("Create transaction failed %v", err)
		return txnID.ID, err
	}

	transactionResponse, err := channel.SendTransaction(tx)
	if err != nil {
		_LOGGER.Fatalf("Send transaction to orderer failed %v", err)
		return txnID.ID, err

	}

	if transactionResponse.Err != nil {
		_LOGGER.Fatalf(" Transaction response from Orderer %v\n", transactionResponse)
		return txnID.ID, err
	}

	_LOGGER.Printf(" Transaction response from Orderer %v\n", transactionResponse)
	return txnID.ID, nil
}
