package gorelic

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	metrics "github.com/yvasiyarov/go-metrics"
	"github.com/yvasiyarov/gorelic"
)

var agent *gorelic.Agent

// Handler for the agent
func Handler(c *gin.Context) {
	startTime := time.Now()
	c.Next()
	if agent != nil {
		agent.HTTPTimer.UpdateSince(startTime)
	}
}

// InitNewrelicAgent creates the new relic agent
func InitNewrelicAgent(license string, appname string, verbose bool) error {

	if license == "" {
		return fmt.Errorf("Please specify NewRelic license")
	}

	agent = gorelic.NewAgent()
	agent.NewrelicLicense = license

	agent.HTTPTimer = metrics.NewTimer()
	agent.CollectHTTPStat = true
	agent.Verbose = verbose

	agent.NewrelicName = appname
	agent.Run()
	return nil
}
