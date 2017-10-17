# Various query method of Hyperledger Fabric 
This document describes the various query method available to know the following information about a block chain network.
1. Block Info - Synopsys of the entire ledger assocaited with a channel with the details to access the top most block
2. Individual block details - Contatents of a given block 
3. Transaction details - Details of a given transaction 

## Details of each query method
### 1. Block Info
    To retrive the ledger details with the top/latest block use the following method of a channel instance. ( Please refer to the GenerateStatistics in the appclient.go file)
    ```go
    import (
        "fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient""
    )

    var channel fab.Channel 
    ....
    blockInfo, err := channel.QueryInfo()

    ```
    The blockInfo consist of the following details 
    
    ```
    type BlockchainInfo struct {
	    Height            uint64 
	    CurrentBlockHash  []byte 
	    PreviousBlockHash []byte 
    }
    ```
    With the currert block hash / previous block hash as input the individual block details could be retrieved. 
### 2. Individual block details
    Individual block details could be retived using following  3 methods . ( Please refer to the GenerateStatistics in the appclient.go file)
    
   ```go
    import (
        "fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient""
    )

    var channel fab.Channel 
    ....
    blockInfo, err := channel.QueryBlockByHash(hasbytes)
    or
    blockInfo, err :=channel.QueryBlock(block number) //Block numbers ranges from 0 to BlockchainInfo.Height -1

    ```
    The structure of the blockInfo object is like the following 
    ```
    type Block struct {
        Header   *BlockHeader   
        Data     *BlockData     
        Metadata *BlockMetadata 
    }
    type BlockHeader struct {
        Number       uint64 
        PreviousHash []byte 
        DataHash     []byte `
    }
    
    type BlockData struct {
        Data [][]byte 
    }

    ```
    The PreviousHash of the BlockHeader could be utilized to traverse back to the another block in the ledger or directly using the block number.
    Regarding the BlockData structure , each line looks like a transaction details , and spliting with a '\n' of each row of BlockData.Data , revels the transaction details (like input params, signatures etc etc. )( Please run the application.go file to generate an example output )

### 3. Transaction details 
    Transaction details of a given transaction id ( this need to be collected while executing the transaction ) could be retrived using the following method
    ```go
     import (
        "fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient""
    )

    var channel fab.Channel 
    ....
    trxnInfo, err := channel.QueryTransaction(trxnId)
    ```
    The structure of the transaction info is 
    ```
    type ProcessedTransaction struct {
    // An Envelope which includes a processed transaction
        TransactionEnvelope *common.Envelope `protobuf:"bytes,1,opt,name=transactionEnvelope" json:"transactionEnvelope,omitempty"`
    // An indication of whether the transaction was validated or invalidated by committing peer
        ValidationCode int32 `protobuf:"varint,2,opt,name=validationCode" json:"validationCode,omitempty"`
    }
    ```
 