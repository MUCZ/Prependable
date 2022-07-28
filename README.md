# Prependable
Docs : https://pkg.go.dev/github.com/mucz/prependable

Prependable is a buffer that support 'Prepend(data []byte)'
method without copying and moving the existing payload inside
the buffer.

It is useful when building networking packets, where each
protocol adds its own headers to the front of the
higher-level protocol header and payload; for example, TCP
would prepend its header to the payload, then IP would
prepend its own, then ethernet.

The larger the (len(payload)/len(header)), the more Prependable
Buffer can show performance benefits.

```
|- reserved  space -|
+-------------------+------------------+------------------+
| prependable bytes |  readable bytes  |  writable bytes  |
|                   |     (CONTENT)    |                  |
+-------------------+------------------+------------------+
|                   |                  |                  |
0      <=      readerIndex   <=   writerIndex    <=     size
```

<p align="center">
  <img src="https://github.com/MUCZ/Prependable/blob/main/report.jpg">
</p>
