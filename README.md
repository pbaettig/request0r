# tests
A list of Tests (see below)

# Test
## id
## numRequests
## concurrency
## targetRequestsPerSecond
## urlSpecs
A list of URLSpec (see below)

## URLSpec
An URLSpec describes the components of an URL. The program will generate however many URLs it needs according to this specifications.
## scheme
## host
## uriComponents
A list of PathComponents that describe the parts of the URI

# PathComponent
## type: string
### value

## type: randomString
### chars
### minLength
### maxLength
### format

## type: integer
### min
### max

## type: httpStatus
### ranges
