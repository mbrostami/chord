# learning puprose

# Chord
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
https://arxiv.org/pdf/1502.06461.pdf

# Distributed Storage 

# Merkle Tree syncronization
### Defenition   
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


# TODO
-[] use https://github.com/grpc/grpc/blob/master/doc/health-checking.md instead of ping  
-[] Virtual nodes   
-[] FIX: sometimes when a node fails, the predecessor of that node, updates it's successor to itself instead of picking next one from successor list!   

# Debug 
```
bootstrap-node# go run cmd/main.go -vvv # deault host:port localhost:10001

node1#  go run cmd/main.go --port 10002 -vvv # default bootstrap node is localhost:10001

node2#  go run cmd/main.go --port 10003 -vvv # default bootstrap node is localhost:10001
.
.
.

```
**Change verbose output in ring.go -> verbose**