# learning puprose

# Chord
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
https://arxiv.org/pdf/1502.06461.pdf

# Storage 
https://pdos.csail.mit.edu/papers/sit-phd-thesis.pdf    
OpenDHT   

# Merkle Tree syncronization
Implementation of merkle tree syncronization   
Consider following rules:  
- A record is a key-value pair of data
- key-value pairs are immutable (data will not be edited)
- One block contains a range of records    
- Block hash is calculated based on records hashes   
- A block hash is a leaf in merkle tree   
To make 2 nodes synced, we need to make sure both have the same data. 
To do that we need to compare data in both nodes periodically.
To do the comparison, the simple solution is to send keys from the first node to the next node and return the missing/additional keys to be transfered later. In large amount of records, this is not an optimized way. So we use Blocks. Each block contains multiple key-value records. Now we can transfer hash of the blocks instead of keys. But still the same issue. How many blocks can we support? What if we have a large amount of blocks? How many records each block should have? And more importantly how both nodes know how to make blocks with exactly same records? Do we need range for each block to be sent?   
What we do, is to make blocks in dynamic size. Considering the insertion timestamp of each record we can calculate a logarithmic block number.  


By using records creation timestamp, we make a dynamic sized blocks to make merkle trees  
First block can contain many records with the short lifetime, as the block number increases  
the number of records in block increases.  
Using the following simple math it's possible to detect the block number out of creation timestamp:  
```
lifetime = now - creation timestamp of a record
blocknumber = round(log2(record lifetime))
```   
This will return an integer as we use that as block index or block number.  
By this way, all new records will go to the head blocks and old data will remain in the old block number. So we would have less block movement for old data.  
e.g.  
```
// a record which is inserted 2 seconds ago
blocknumber = round(log2(2)) = 1

// a record which is inserted 15 seconds ago
// a record which is inserted 16 seconds ago
blocknumber = round(log2(10)) = 4

// a record which is inserted 2100 seconds ago
blocknumber = round(log2(2100)) = 11

// a record which is inserted 2800 seconds ago
blocknumber = round(log2(2800)) = 11
.
.
.

```


# TODO
-[] use https://github.com/grpc/grpc/blob/master/doc/health-checking.md instead of ping  
-[] Virtual nodes  


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