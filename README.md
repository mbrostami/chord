# learning puprose

# Chord
https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   


# Storage 
https://pdos.csail.mit.edu/papers/sit-phd-thesis.pdf    
OpenDHT   



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