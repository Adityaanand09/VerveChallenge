To prevent de-duplication when two instances of the
server is running, i have chosen to use Redis where we can store the
unique ids along with a TTL of 1 minute.

I have used a sorted set in this case to where the score of each
key is going to be its currentTimestamp and ids older than 1 minute
are removed via a function which runs every 1 minute in a separate go-routine.

# Reason for using Redis
1. Since it is an in-memory database, and it provides extremely fast 
read and write operations with very low latency.Since we will be reading number of unique ids everytime
therefore, we need the throughput to be high and latency to be low.
2. Since operations like `ZADD` are atomic in redis, therefore we don't need to do any kind of locking
which might be required in the SQL DBs like mysql that can lead to an increase in the latency.