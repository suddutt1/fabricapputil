# Various query method of Hyperledger Fabric 

This document describes the various query method available to know the following information about a block chain network.

1. Block Info - Synopsys of the entire ledger assocaited with a channel with the details to access the top most block
2. Individual block details - Contatents of a given block 
3. Transaction details - Details of a given transaction 

## Details of each query method

### 1. Block Info
To retrive the ledger details with the top/latest block use the following method of a channel instance. 
( Please refer to the GenerateStatistics in the appclient.go file)

    ```go
    
    import (
        fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
    )

    var channel fab.Channel 
    ...
    blockInfo, err := channel.QueryInfo()
    ```
        
    
### 2. Individual block details

   

### 3. Transaction details 
    
 
