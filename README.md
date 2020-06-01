# learning puprose

# Chord
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
https://arxiv.org/pdf/1502.06461.pdf

# Storage 
https://pdos.csail.mit.edu/papers/sit-phd-thesis.pdf    
OpenDHT   

# Merkle Tree
Customized implementation of merkle tree   
By using records creation timestamp, we make a dynamic sized blocks to make merkle trees  
First block can contain many records with the short lifetime, as the block number increases  
the number of records in block increases.  
Using the following simple math it's possible to detect the block number out of creation timestamp:  
```
lifetime = now - record creation timestamp
round(log2(record lifetime))
```   
This will return an integer as we use that as block index or block number.  
By this way, all new records will go to the first blocks and old data will remain in the old block number.  
e.g.  
```
// a record which is inserted 2 seconds ago
blocknumber = round(log2(2)) = 1

// a record which is inserted 16 seconds ago
blocknumber = round(log2(10)) = 4

// a record which is inserted 2100 seconds ago
blocknumber = round(log2(2100)) = 11

// a record which is inserted 2800 seconds ago
blocknumber = round(log2(2)) = 11
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