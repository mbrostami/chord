# Chord  
Chord protocol implemented based on https://pdos.csail.mit.edu/papers/ton:chord/paper-ton.pdf   
Dstore implemented based on https://pdos.csail.mit.edu/papers/sit-phd-thesis.pdf   


# TODO
[] Bootstrapping issue
[] add channel to chord to send notification to channel, dstore will update storage based on these changes
[] validation on username
[] identify who want to store data? (digital signature?)
[] prevent duplicate username
[] improvement part for node valuntary leave
[] change sha256 to sha1 
[] add static number of virtual hosts to chord


# Information
Using chord for routing protocol + lookup nodes which stores specific keys   
Using dstore which is self implemented protocol to store and retrieve data from network    
Using chord vhosts to balance network, it's staticlly defined   
