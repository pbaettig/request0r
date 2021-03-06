# request0r
A simple utility to generate a number of requests to random URLs that follow a user-provided specification. The URL format and test specifics are described in a YAML file (see [Test Examples](https://github.com/pbaettig/request0r#test-examples)).
This is primarily a pet project to help me learn Go, but maybe it'll be helpful to others as well.


## Config file format
### tests
A list of Tests (see below). Take a look at [Test Examples](https://github.com/pbaettig/request0r#test-examples).

### Test
#### id
String: Name of the test (required)
#### numRequests
Integer: Total number of requests to execute (required)
#### concurrency
Integer: Number of workers executing requests in parallel (required)
#### targetRequestsPerSecond
Integer: Target rate of requests per second. If left empty or set to 0 no throttling will be performed.
#### urlSpecs
A list of URLSpec that define the URLs under test (required)

### URLSpec
An URLSpec describes the components of an URL. The program will generate however many URLs it needs according to this specifications.
#### scheme
String: Either http or https
#### host
String: The host targeted by the test. If required a custom port can be specified as part of it.
#### uriComponents
A list of `PathComponent`s that describe the parts of the URI

### PathComponent
A URI consists of a number of PathComponents that are joined using "/". There are different `PathComponent`s available:
#### type: string
A static string value.
##### value
String: Value of the component (required)
#### type: randomString
A random string value.
##### chars
String: The characters that can be used to generate the string. (required)

#### format
String: If specified this format string will be used for generating the string. The format for the string literal is as follows 

```%<minLength>,<maxLength>s``` e.g. ```%1,4s``` which will produce a random string consisting of `chars` with a length of between 1-4 characters.
#### type: integer
A random integer value
##### min
Integer: minimum value (required)
##### max
Integer: maximum value (required)

#### type: httpStatus
A valid HTTP status code
##### ranges
A list of acceptable ranges for the generated code, e.g. 200, 300, 500 (required)

## Test examples
Run a test called `user-details`. Execute 400 requests from 10 workers with at most 20 requests per second. The generated URLs will look something like this: https://user-mgmt.acme.com/user/user-32dd-14a1-d7f8-d322/details
```yaml
tests:
  - id: user-details
    numRequests: 400
    concurrency: 10
    targetRequestsPerSecond: 20
    urlSpecs:
      - scheme: https
        host: user-mgmt.acme.com
        uriComponents:
          - type: string
            value: user
          - type: randomString
            chars: abcdef0123456789
            format: "user-%4,4s-%4,4s-%4,4s-%4,4s"  
          - type: string
            value: details
```
