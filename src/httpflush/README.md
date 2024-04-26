<pre>
██╗  ██╗████████╗████████╗██████╗ ███████╗██╗     ██╗   ██╗███████╗██╗  ██╗
██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔════╝██║     ██║   ██║██╔════╝██║  ██║
███████║   ██║      ██║   ██████╔╝█████╗  ██║     ██║   ██║███████╗███████║
██╔══██║   ██║      ██║   ██╔═══╝ ██╔══╝  ██║     ██║   ██║╚════██║██╔══██║
██║  ██║   ██║      ██║   ██║     ██║     ███████╗╚██████╔╝███████║██║  ██║
╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═╝     ╚══════╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝
</pre>

| **`root version`** | **`package version`** |
| ------------------ | --------------------- |
| `v0.1.4`           | `v0.1.0`              |

## Usage

```go
import "github.com/tcodes/go/src/httpflush"

// the writer you have in your handler
var w http.ResponseWriter

// wrap w with httpflush.MaxSize
m := httpflush.MaxSize{
		Max:    1024,
		Writer: w,
	}

// use m as the writer moving forward, pass it to other functions
```

## see also

[max_size.go](https://github.com/tcodes0/go/tree/main/src/httpflush/max_size.go)<br/>
[max_size_test.go](https://github.com/tcodes0/go/tree/main/src/httpflush/test/max_size_test.go)<br/>