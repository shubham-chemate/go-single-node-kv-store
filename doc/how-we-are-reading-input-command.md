## How we are reading input command?

Our input reading is highly inspired by Redis RESP protocol

here is what Claude says about RESP
/*
### Redis RESP Protocol

RESP (REdis Serialization Protocol) is a simple, text-based protocol used by Redis for client-server communication.

#### Key Components

##### Data Types

RESP uses prefixes to identify data types:

- **`*`** - Arrays (e.g., `*3` means array with 3 elements)
- **`$`** - Bulk Strings (e.g., `$6` means string of 6 bytes)
- **`+`** - Simple Strings (e.g., `+OK`)
- **`-`** - Errors (e.g., `-ERR message`)

##### Message Format

Commands are sent as **arrays of bulk strings**. Each line ends with `\r\n`.

**Example**: `SET pin 414103`
```
*3\r\n$3\r\nSET\r\n$3\r\npin\r\n$6\r\n414103\r\n

*3\r\n           # Array with 3 elements
$3\r\n           # First element: 3-byte string
SET\r\n          # The command
$3\r\n           # Second element: 3-byte string
pin\r\n          # The key
$6\r\n           # Third element: 6-byte string
414103\r\n       # The value
```

##### Response Format

- Success: `+OK\r\n` or `$6\r\nvalue1\r\n`
- Error: `-ERR message\r\n`
- Integer: `:42\r\n`

This protocol is simple, human-readable, and easy to parse.

*/

- For simplicity purposes, we are treating everything as string (even the size of arrays and bulk strings)
- We convert the value that we get into integer in later stage
- We have \r\n as a delimiter, the characters between two delimeter is our token
- We have configured max allowed token size, we don't read more that max allowed token size and returns error if your token size is greather than this we will return error and close the connection
- We allow three types of commands
    - SET key val ttl : this sets the value for key in our KV store, if key already present, it will override the value. ttl is optional
    - GET key : this will return value of key from KV store, if key is not present, it will return -1
    - DEL key : this will remove key from KV store