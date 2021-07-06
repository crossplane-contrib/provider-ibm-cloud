# Event Streams Admin V1 API examples

The Event Streams Admin V1 API allows to provision and manage topics.

# Setting Partitions and PartitionCount

If Partitions and PartitionCount are both not set they will be initialized to 1. If Partitions are set and PartitionCount is not set, PartitionCount will be initialized to what Partitions is. If Partitions is not set and PartitionCount is set then Partitions will be initialized to what PartitionCount is. 

# Updating the number of partitions

The number of partitions can only be increased. The number of partitions will be updated to what PartitionCount is. 

# Troubleshooting

The number of partitions are only updated if PartitionCount is greater than the current number of partitions.

If PartitionCount is updated to be less than the current number of partitions no error event will be returned. The controller will continually attempt to update the number of partitions, but the number of partitions will not be updated. 