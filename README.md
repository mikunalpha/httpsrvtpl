# httpsrvtpl
A HTTP Service Project Structure Template In Go.

## Dependencies
You can replace below tools and packages if you want.

### Tools
**Package Management**  
  - `github.com/Masterminds/glide`  

### Packages
**Logger**  
  - `github.com/sirupsen/logrus`  

**JSON**  
  - `github.com/json-iterator/go`  

**Router**  
  - `github.com/gin-gonic/gin`  

**Authentication**  
  - `github.com/dgrijalva/jwt-go`  

**Command Line**  
  - `github.com/urfave/cli`

**Testing**  
  - `github.com/stretchr/testify`  

## Test
Run below commands to detect data race.
```
go test -race ./...
```
Run below commands to see coverprofle.
```
go test -coverprofile=c.out ./... && gg tool cover -html=c.out
```

## Build
Run below commands to build executable binary.
```
go build -tags=jsoniter -ldflags="-s -w"
```