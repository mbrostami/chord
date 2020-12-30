# Merkle Tree syncronization [TODO]
## Defenition   
Source Time = start time of a syncronization, to have same block numbers while replication  
Master Block = contains a tree with multiple blocks + root hash + min + max   
Tree = a merkle tree structure using blocks   
Block = leaf node of a merkle tree might contain multiple records but only one hash   
Replication = the service to replicate data, using 3 above   

## Merkle Tree Structure
Consider following rules:  
- A record is a key-value pair of data
- key-value pairs are immutable (data will not be edited)
- One block contains a range of records    
- Block hash is calculated based on records hashes   
- A block hash is a leaf in merkle tree   
To make 2 nodes synced, we need to make sure both have the same data. To do that we need to compare data in both nodes periodically. To do the comparison, the simple solution is to hash each record and send keys from the first node to the next node and return the missing/additional keys to be transfered later. In situations with large amount of records, this is not an optimized way. So we use Blocks. Each block contains multiple key-value records. Now we can transfer hash of the blocks instead of keys. But still the same issue. How many blocks shall we support? What if we have a large amount of blocks? How many records each block should have? And more importantly how both nodes know how to make blocks with exactly same records? Do we need to send the ranges(min-max) of the hashes for each block?   
What we do, is to make blocks in dynamic size. Considering the insertion timestamp of each record we can calculate a logarithmic block number.   

By using record's creation timestamp, we make a dynamic sized blocks to make merkle trees. First block can contain a few records with the short lifetime, as the block number increases the number of records in block increases. Using the following simple math it's possible to detect the block number out of creation timestamp:  
```
SourceTime := time.Now()
lifetime := SourceTime - creation timestamp of a record
blocknumber = round(log2(record lifetime))
```   

This will return an integer as we use that as block index or block number.  
By this way, all new records will go to the head blocks and old data will remain in the old block number. So we would have less block movement for old data.  


## Master Block
Each Master Block contains one merkle tree of smaller blocks (leafs). Number of MasterBlocks is exactly the same as number of replications we want in the application.  
e.g.   
Imagine we have 5 nodes: a,b,c,d,e  
We want to have 3 copy of data in different nodes. Sync starts in node c. To make 3 replicas possible, Node c must create 2 master block with data between (a and b] and (b and c] to send them to successor node d.   
MasterBlock1 = (a and b]   
MasterBlock2 = (b and c]   
In situations like node failure, master block will be helpful to easily detect the missing data in successor. In this scenario, if node d fails, syncronization process will start between node c and node e. As node e has already received data (b and c] from node d, so the only part which should be added to node e, is (a and b]. This is why we are using Master Blocks.  

## Migration
- node n makes merkle trees with local database  
- node n sends roothashes+treeranges+sourcetime to node s  
- node s makes merkle trees with local database  
- node s receives n's roothashes and compares with local ones  
- if diff found, it sends back all leafs sorted (block hashes) to node n  
- node n sorts local leafs (block hashes) and compares one by one (queue workers)  
- for each diff, node n extracts records of that block, sorts, and starting from newest one, sends to node s  
- node s receives new data and stores in db, recalculate that block hash and returns back to node n  
- node n checks if new block hash received from node s is matched with local block hash, will skip this block, otherwise will continue sending data  
**NOTE** stream - stream grpc connection with queue workers  
- 
