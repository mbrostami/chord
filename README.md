# learning puprose

# Chord
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
https://arxiv.org/pdf/1502.06461.pdf

# Storage 
https://pdos.csail.mit.edu/papers/sit-phd-thesis.pdf    
OpenDHT   

# Merkle Tree
Custom implementation of merkle tree   
By using records creation timestamp, we make a dynamic sized blocks to finally make merkle tree  
First block can contain many records with the short lifetime, as the block number increases  
the number of records in block increases and holds older data. 

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