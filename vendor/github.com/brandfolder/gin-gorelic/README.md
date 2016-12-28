GoRelic
=======

NewRelic middleware for gin-gonic framework.

## Usage

```go
import(
	"github.com/gin-gonic/gin"
	"github.com/brandfolder/gin-gorelic"
)

func main(){
	g := gin.Default()

	gorelic.InitNewrelicAgent("YOUR_NEWRELIC_LICENSE_KEY", "YOUR_APPLICATION_NAME", true)
	g.Use(gorelic.Handler)

	g.Run()
}
```

## Authors

* [Jason Waldrip](http://github.com/jwaldrip)
* [Yuriy Vasiyarov](http://github.com/yvasiyarov) [original martini gorelic]
