# Snowflake

Thanks to Twitter’s internal technology, the snowflake algorithm can be spread and widely used today because it has several characteristics.

* It can satisfy the non-repetitive ID in the high concurrent distributed system environment.
* High production efficiency.
* Based on timestamps, ordered increments are guaranteed.
* No dependencies on third-party libraries or middleware.
* The generated id is sequential and unique.

## How It Works
Each time you generate an ID, it works, like this.

A timestamp with millisecond precision is stored using 41 bits of the ID.
Then the NodeID is added in subsequent bits.
Then the Sequence Number is added, starting at 0 and incrementing for each ID generated in the same millisecond. If you generate enough IDs in the same millisecond that the sequence would roll over or overfill then the generate function will pause until the next millisecond.
The default Twitter format shown below.


| 1 Bit Unused | 41 Bit Timestamp |  10 Bit NodeID  |   12 Bit Sequence ID |


Using the default settings, this allows for 4096 unique IDs to be generated every millisecond, per Node ID.

## Result


|1000   | 1ms  |
|-------|------|
|10000  | 7ms  |
|100000 | 90ms |

## Reference
https://betterprogramming.pub/implementing-snowflake-algorithm-in-golang-c1098fdc73d0