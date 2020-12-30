Going to be a distributed anonymous cli chat

# Chord
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
https://arxiv.org/pdf/1502.06461.pdf

# Distributed Storage 

### Storing on the remote node
In order to store data in the network, the source node needs to calculate the hash of data and calls Store method of the hash's successor node. The new node will store data if the local calculation of the data is in range (predecessor, new node] 

### Join initial download
Whenever a node joins the network, it must connect to the successor and immediately downloads the range of data between predecessor and new node (data E (predecessor, node]) from the successor. After downloading and storing this range of data, the node can be considered as a joint node in ring hash.   

### Sync data
In order to sync data with node A and its successors (B, C, D), it depends on the number of replication we need in the network. Currently, we have 2 replicas of each data. There is another document (REPLICATION.md) that is a more complex and efficient way of implementing this, but for now, we keep this as simple as possible.   

Node A fetches data from the local database with range scan (predecessor, node A]. This range needs to be transferred to the successor to make a replica. But we can't transfer all data all the time. So node A makes a root hash of all the data in that range. Then sends this root hash and the range's first/last identifiers (predecessor, node A) to the successor. Successor fetches given range from the database, makes a new root hash, compares with the given root hash, if it's the same it will return nil, otherwise it will send back all the data it has to the predecessor (node A).   
Node A receives the response and checks data record by record to see what data is missing in local and what data is missing in the successor. Then it will store the missing data in the local and successor node. 



# TODO
-[] use https://github.com/grpc/grpc/blob/master/doc/health-checking.md instead of ping  
-[] Virtual nodes   
-[] FIX: sometimes when a node fails, the predecessor of that node, updates its successor to itself instead of picking the next one from the successor list!   

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