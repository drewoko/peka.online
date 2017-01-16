package core

import (
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"log"
)

func StartWeb(config *Config, db *DataBase) {

	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovering WEB Instance", r)
		}
	}()

	ginInst := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	if config.Dev {
		ginInst.Use(static.Serve("/", static.LocalFile(config.Static, true)))
	} else {
		ginInst.Use(static.Serve("/", BinaryFileSystem("static")))
		ginInst.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	ginInst.NoRoute(redirect)

	ginInst.GET("/stats", func(c *gin.Context) {

		statistics := db.GetStatistics()

		var resp []StatisticsResponseStruct

		for i := 0; i < len(statistics); i = i + 2 {

			ggStat := statistics[i]
			pekaStat := statistics[i+1]

			resp = append(resp, StatisticsResponseStruct{
				Time:          ggStat["externalTime"].(string),
				GGMessages:    ggStat["messageCount"].(int),
				PekaMessages:  pekaStat["messageCount"].(int),
				GGUniqUsers:   ggStat["uniqUsers"].(int),
				PekaUniqUsers: pekaStat["uniqUsers"].(int),
			})
		}

		c.JSON(200, resp)
	})

	ginInst.GET("/stats/aggregate", func(c *gin.Context) {
		statistics := db.GetAggregateStatistics()

		c.JSON(200, statistics)
	})

	ginInst.Run(":" + config.Port)
}

type StatisticsResponseStruct struct {
	Time          string
	GGMessages    int
	PekaMessages  int
	GGUniqUsers   int
	PekaUniqUsers int
}

func redirect(c *gin.Context) {
	c.Redirect(301, "/index.html")
}
