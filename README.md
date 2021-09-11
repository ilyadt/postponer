# Simple http service for delayed jobs

## Request

```
/add?queue=<queue>&body=<body>&delay=<delay>
```

`<queue>`(string) - arbitrary queue
`<body>`(string)  - arbitrary queue body
`<delay>`(int) - time in seconds to wait before proceed


## Principle

At least once.

Not earlier than delay.

FIFO not guaranteed.
