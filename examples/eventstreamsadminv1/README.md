# Event Streams Admin V1 API examples

The Event Streams Admin V1 API allows to provision and manage topics.

# Setting Partitions and PartitionCount

If Partitions and PartitionCount are both not set they will be initialized to 1. If Partitions are set and PartitionCount is not set, PartitionCount will be initialized to what Partitions is. If Partitions is not set and PartitionCount is set then Partitions will be initialized to what PartitionCount is. 

# Updating the number of partitions

The number of partitions can only be increased. The number of partitions will be updated to what PartitionCount is. 