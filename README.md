# Web Server with Simple HTTP Protocol

### HTTP Messages

TritonHTTP follows the [general HTTP message format](https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages). And it has some further specifications:

- HTTP version supported: `HTTP/1.1`
- Request method supported: `GET`
- Response status supported:
  - `200 OK`
  - `400 Bad Request`
  - `404 Not Found`
- Request headers:
  - `Host` (required)
  - `Connection` (optional, `Connection: close` has special meaning influencing server logic)
  - Other headers are allowed, but won't have any effect on the server logic
- Response headers:
  - `Date` (required)
  - `Last-Modified` (required for a `200` response)
  - `Content-Type` (required for a `200` response)
  - `Content-Length` (required for a `200` response)
  - `Connection: close` (required in response for a `Connection: close` request, or for a `400` response)
  - Response headers should be written in sorted order for the ease of testing

### Server Logic

When to send a `200` response?
- When a valid request is received, and the requested file can be found.

When to send a `404` response?
- When a valid request is received, and the requested file cannot be found or is not under the doc root.

When to send a `400` response?
- When an invalid request is received.
- When timeout occurs and a partial request is received.

When to close the connection?
- When timeout occurs and no partial request is received.
- When EOF occurs.
- After sending a `400` response.
- After handling a valid request with a `Connection: close` header.

When to update the timeout?
- When trying to read a new request.

What is the timeout value?
- 5 seconds.

## Usage

Install the `httpd` command to a local `bin` directory:
```
make install
ls bin
```

Check the command help message:
```
bin/httpd -h
```

An alternative way to run the command:
```
go run cmd/httpd/main.go -h
```

## Testing

### Sanity Checking

We provide 2 simple examples for your sanity checking.

First you could run an example with the default server:
```
make run-default
```

This example uses the Golang standard library HTTP server to serve the website, and it doesn't rely on your implementation of TritonHTTP at all. So you shall be able to run it with the starter code right away. Open the link from output in a browser, and you shall see a test website.

Once you have a working implementation of TritonHTTP, you could run another example:
```
make run-tritonhttp
```

Again, you could use a browser to check the test website served.

### Unit Testing

Unit tests don't involve any networking. They check the logic of the main parts of your implementation.

To run all the unit tests:
```
make unit-test
```

### End-to-End Testing

End-to-end tests involve runing a server locally and testing by communicating with this server.

To run all the end-to-end tests:
```
make e2e-test
```

### Manual Testing

For manutal testing, we recommend using `nc`.

In one terminal, start the TritonHTTP server:
```
go run cmd/httpd/main.go -port 8080 -doc_root test/testdata/htdocs
```

In another terminal, use `nc` to send request to it:
```
cat test/testdata/requests/single/OKBasic.txt | nc localhost 8080
```

You'll see the response printed out. And you could look at your server's logging to debug.
