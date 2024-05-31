<pre>
██╗  ██╗████████╗████████╗██████╗ ███████╗██╗     ██╗   ██╗███████╗██╗  ██╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔════╝██║     ██║   ██║██╔════╝██║  ██║
███████║   ██║      ██║   ██████╔╝█████╗  ██║     ██║   ██║███████╗███████║
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══╝  ██║     ██║   ██║╚════██║██╔══██║
██║  ██║   ██║      ██║   ██║     ██║     ███████╗╚██████╔╝███████║██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝     ╚══════╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝
</pre>

## Versions

| **`root version`** | **`package version`** |
| ------------------ | --------------------- |
| `v0.1.4`           | `v0.1.0`              |

## Description

Calls `Flush()` on the `http.ResponseWriter` when the size of the response body exceeds a set max size in bytes.
Flushing the writer sends the buffered bytes to the client.

## Usage

```go
import "github.com/tcodes/go/httpflush"

// the writer you have in your handler
var w http.ResponseWriter

// wrap w with httpflush.MaxSize
m := httpflush.MaxSize{
		Max:    1024, // set your desired size
		Writer: w,
	}

// use m as the writer moving forward, pass it to other functions
```

## See also

[max_size.go](https://github.com/tcodes0/go/tree/main/httpflush/max_size.go)<br/>
[max_size_test.go](https://github.com/tcodes0/go/tree/main/httpflush/httpflush_test/max_size_test.go)<br/>
