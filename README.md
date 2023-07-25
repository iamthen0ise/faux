# Faux - A Customizable HTTP Server

Faux is a customizable HTTP server written in Go. It acts like HTTPBIN but with additional functionality to support user-defined routes with predefined HTTP methods, paths, status codes, response headers and bodies.

## Features

- Add custom routes with specified HTTP methods, paths, status codes, response bodies and headers.
- Support for magic routes, which allow dynamic generation of responses based on the request.
- Detailed and customizable logging with optional color output.

## Usage

1. Clone this repository:

```bash
git clone https://github.com/yourusername/faux.git
```

Build the project:

```bash
    cd faux
    go build
```


Run the server:

```bash
./faux
```

## Customization

You can customize the behavior of Faux by providing a JSON file with routes and using command line flags when running the server.
Routes

Routes are defined in a JSON file with the following structure:

```json
[
	{
		"Path": "/custom",
		"Method": "GET",
		"StatusCode": 200,
		"ResponseHeaders": {
			"Content-Type": "application/json"
		},
		"ResponseBody": "{\"message\":\"Hello, world!\"}"
	}
]
```

You can specify as many routes as you want in the array. The Path and Method fields are required, but ResponseHeaders and ResponseBody are optional.
## Magic Routes

Magic routes allow dynamic responses based on the request. For example, a GET request to /status/200/?response_headers={...}&response_body={...} will return an HTTP 200 response with the specified headers and body. POST and PUT requests can specify headers and body in the request payload.
## Logging

By default, Faux logs the time, method, status code, path and response time for each request. You can customize this by using a format template with the -format flag when running the server. For example:

```bash
./faux -format="{{.Time}} {{.Method}} {{.StatusCode}} {{.Path}} {{.ResponseTime}}"
```
You can disable color output with the -no-color flag:

```bash
./faux -no-color
```
## License
