## How we are reading input command?

Our input reading is highly inspired by Redis RESP protocol

here is what Claude says about RESP  
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

**Example**: this is the command in redis-cli `SET pin 414103`
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

---

## Process

- For simplicity purposes, we are treating everything as string (even the size of arrays and bulk strings)
- We convert the size value that we get into integer in later stage
- We have \r\n as a delimiter, the characters between two delimeter is our token
- We have configured max allowed token size, we don't read more that max allowed token size and returns error if your token size is greather than this we will return error and close the connection
- We have basic validations on the token that we are reading like for array size, there should be * at the start, for bulk string size, there should be $ at the start, every token must end with \r\n. if this basic checking fails we will return error, the redis-cli takes care of all of this, so from user POV, it's not that hard
- We allow three types of commands
    - SET key val ttl : this sets the value for key in our KV store, if key already present, it will override the value. ttl is optional
    - GET key : this will return value of key from KV store, if key is not present, it will return -1
    - DEL key : this will remove key from KV store
- after reading the entire array, we will give to to command processor to process, the output of command processor is returned to the client (+OK or -ERR)

### Several approaches for reading methods

- Initially though about several approaches for reading methods
- We are using byte by byte read while reading size of the input array / bulk string, since we don't want to allow infinite size for that
- For reading command, keys and values, we are using buffer of size that we already read and ReadFull method
- Apart from size, byte by byte reading is not convenient for us to implement, for size reading it is flexible, simple and efficient in our case (initially we were using byte by byte approach for command reading as well but later realized that bulk bytes reading using buffer is much better approach)
- We also though about buffer approach, but it was creating lot of mess and it will be like reinventing the wheel
- For scanner approach, we find out that it has default max token size of 64KB, it is configurable but scanner is designed for more of a like line by line reading approach
- The only catch with ReadFull is what if client doesn't agree on it's predefined input command size, but we are trusting redis protocol for that and we have checks in place
- The ReadAll will be pretty bad in out case since it will read whatever it can!