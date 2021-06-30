# EQ Works - Backend Track

1. A string value named "key" was added to the counters structure. The value of key contains the content category and the timestamp. An array of counters stores a certain number of minutes worth of counters, where earlier indexes (i.e. index 0, 1...) indicate more recent data.

2. The mock store uses both memory and a text file. Upon startup, the previously saved counters (if any) are retrieved from 'mockstore.txt' and saved to an array of counters, 'c'. The stats page queries the in-memory counter array.

3. This goroutine writes to the mockstore.txt text file and then sleeps for 5 seconds.

4. The global rate-limit was created using a request counter that refreshes every minute. This global limit can be changed by modifying the 'max' value.